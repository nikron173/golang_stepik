package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

type TestCase struct {
	Request *SearchRequest
	Result  *SearchResponse
	IsError bool
}

// type CheckoutResult struct {
// 	Status  int
// 	Balance int
// 	Err     string
// }

// код писать тут
func SearchServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "application/json")
	if r.Method != http.MethodGet {
		http.ErrNotSupported.Error()
		return
	}
	param := new(SearchRequest)
	param.OrderField = "Name"
	param.Limit = 25

	var err error

	if query := r.FormValue("query"); query != "" {
		param.Query = query
	}
	if orderField := r.FormValue("order_field"); orderField != "" {
		if orderField == "Id" || orderField == "Age" || orderField == "Name" {
			param.OrderField = orderField
		} else {
			w.WriteHeader(http.StatusBadRequest)
			jsonErr, _ := json.Marshal(SearchErrorResponse{
				Error: "ErrorBadOrderField",
			})
			fmt.Fprintf(w, "%s", jsonErr)
			return
		}
	}
	if orderByStr := r.FormValue("order_by"); orderByStr != "" {
		if param.OrderBy, err = strconv.Atoi(orderByStr); err != nil || param.OrderBy < -1 || param.OrderBy > 1 {
			w.WriteHeader(http.StatusBadRequest)
			jsonErr, _ := json.Marshal(SearchErrorResponse{
				Error: "Error parameter order by",
			})
			fmt.Fprintf(w, "%s", jsonErr)
			return
		}
	}
	if limitStr := r.FormValue("limit"); limitStr != "" {
		if param.Limit, err = strconv.Atoi(limitStr); err != nil || param.Limit < 0 {
			w.WriteHeader(http.StatusBadRequest)
			jsonErr, _ := json.Marshal(SearchErrorResponse{
				Error: "Error parameter limit",
			})
			fmt.Fprintf(w, "%s", jsonErr)
			return
		}
	}
	if offsetStr := r.FormValue("offset"); offsetStr != "" {
		if param.Offset, err = strconv.Atoi(offsetStr); err != nil || param.Offset < 0 {
			w.WriteHeader(http.StatusBadRequest)
			jsonErr, _ := json.Marshal(SearchErrorResponse{
				Error: "Error parameter offset",
			})
			fmt.Fprintf(w, "%s", jsonErr)
			return
		}
	}

	file, err := os.Open("dataset.xml")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonErr, _ := json.Marshal(SearchErrorResponse{
			Error: "Error internal server",
		})
		fmt.Fprintf(w, "%s", jsonErr)
		return
	}

	defer file.Close()
	decoder := xml.NewDecoder(file)
	users := make([]User, 0)
	i := 0

	for t, _ := decoder.Token(); t != nil; t, _ = decoder.Token() {
		switch se := t.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "id":
				{
					users = append(users, User{})
					decoder.DecodeElement(&users[i].Id, &se)
				}
			case "age":
				{
					decoder.DecodeElement(&users[i].Age, &se)
				}
			case "first_name":
				{
					decoder.DecodeElement(&users[i].Name, &se)
				}
			case "last_name":
				{
					lastName := ""
					decoder.DecodeElement(&lastName, &se)
					users[i].Name = users[i].Name + " " + lastName
				}
			case "gender":
				{
					decoder.DecodeElement(&users[i].Gender, &se)
				}
			case "about":
				{
					decoder.DecodeElement(&users[i].About, &se)
					i++
				}
			}
		}
	}

	findUsers := make([]User, 0)

	if len(param.Query) == 0 {
		findUsers = append(findUsers, users...)
	} else {
		for _, user := range users {
			if strings.Contains(user.About, param.Query) ||
				strings.Contains(user.Name, param.Query) {
				findUsers = append(findUsers, user)
			}
		}
	}

	if param.OrderBy != 0 {
		switch param.OrderField {
		case "Id":
			{
				sort.Slice(findUsers, func(i, j int) bool {
					if param.OrderBy == -1 {
						return findUsers[i].Id > findUsers[j].Id
					}
					return findUsers[i].Id < findUsers[j].Id
				})
			}
		case "Age":
			{
				sort.Slice(findUsers, func(i, j int) bool {
					if param.OrderBy == -1 {
						return findUsers[i].Age > findUsers[j].Age
					}
					return findUsers[i].Age < findUsers[j].Age
				})
			}
		case "Name":
			{
				sort.Slice(findUsers, func(i, j int) bool {
					if param.OrderBy == -1 {
						return findUsers[i].Name > findUsers[j].Name
					}
					return findUsers[i].Name < findUsers[j].Name
				})
			}
		}

	}

	// response := &SearchResponse{}

	if param.Offset >= len(findUsers) {
		findUsers = make([]User, 0)
		w.WriteHeader(http.StatusOK)
		// response.Users = findUsers
		jsonFindUsers, _ := json.Marshal(findUsers)
		// jsonResponse, _ := json.Marshal(response)
		fmt.Fprintf(w, "%s", jsonFindUsers)
		return
	}

	if param.Limit > 0 && param.Limit <= len(findUsers) && (param.Offset+1) > 0 && (param.Offset+1) <= len(findUsers) {
		if param.Limit+param.Offset > len(findUsers) {
			findUsers = findUsers[param.Offset:]
		} else {
			findUsers = findUsers[param.Offset : param.Offset+param.Limit]
			// response.NextPage = true
		}
	}

	// response.Users = findUsers
	w.WriteHeader(http.StatusOK)
	// jsonResponse, _ := json.Marshal(response)
	jsonFindUsers, _ := json.Marshal(findUsers)
	fmt.Fprintf(w, "%s", jsonFindUsers)
}

func SearchServerErrorToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "application/json")
	if r.Method != http.MethodGet {
		http.ErrNotSupported.Error()
		return
	}

	_, err := r.Cookie("AccessToken")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		jsonErr, _ := json.Marshal(SearchErrorResponse{
			Error: "Bad AccessToken",
		})
		fmt.Fprintf(w, "%s", jsonErr)
		return
	}
}

func SearchServerErrorUnpackJson(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "application/json")
	if r.Method != http.MethodGet {
		http.ErrNotSupported.Error()
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, "{\"%s\":}", "Users")
}

func SearchServerErrorOkUnpackJson(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "application/json")
	if r.Method != http.MethodGet {
		http.ErrNotSupported.Error()
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{\"%s\":}", "Users")
}

func SearchServerErrorInternalServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "application/json")
	if r.Method != http.MethodGet {
		http.ErrNotSupported.Error()
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	jsonErr, _ := json.Marshal(SearchErrorResponse{
		Error: "SearchServer fatal error",
	})
	fmt.Fprintf(w, "%s", jsonErr)
}

func SearchServerErrorUnknowStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "application/json")
	if r.Method != http.MethodGet {
		http.ErrNotSupported.Error()
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	jsonErr, _ := json.Marshal(SearchErrorResponse{
		Error: "Error access database",
	})
	fmt.Fprintf(w, "%s", jsonErr)
}

func SearchServerErrorTimeout(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "application/json")
	if r.Method != http.MethodGet {
		http.ErrNotSupported.Error()
		return
	}
	w.WriteHeader(http.StatusRequestTimeout)
	time.Sleep(time.Second * 100)
	return
}

func TestFindUser(t *testing.T) {
	cases := []TestCase{
		TestCase{
			Request: &SearchRequest{
				Limit:  5,
				Offset: 0,
			},
			Result: &SearchResponse{
				Users:    []User{},
				NextPage: false,
			},
			IsError: false,
		},
		TestCase{
			Request: &SearchRequest{
				Limit:  -1,
				Offset: 0,
			},
			Result: &SearchResponse{
				Users:    []User{},
				NextPage: false,
			},
			IsError: false,
		},
		TestCase{
			Request: &SearchRequest{
				Limit:  10,
				Offset: 30,
			},
			Result: &SearchResponse{
				Users:    []User{},
				NextPage: false,
			},
			IsError: false,
		},
		TestCase{
			Request: &SearchRequest{
				Limit: 30,
			},
			Result: &SearchResponse{
				Users:    []User{},
				NextPage: false,
			},
			IsError: false,
		},
		TestCase{
			Request: &SearchRequest{
				Offset: -1,
			},
			Result: &SearchResponse{
				Users:    []User{},
				NextPage: false,
			},
			IsError: true,
		},
		TestCase{
			Request: &SearchRequest{
				OrderField: "Kek",
			},
			Result: &SearchResponse{
				Users:    []User{},
				NextPage: false,
			},
			IsError: true,
		},
		TestCase{
			Request: &SearchRequest{
				OrderField: "Kek",
			},
			Result: &SearchResponse{
				Users:    []User{},
				NextPage: false,
			},
			IsError: true,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	for caseNum, item := range cases {
		c := &SearchClient{
			AccessToken: "",
			URL:         ts.URL,
		}

		result, err := c.FindUsers(*item.Request)

		if err == nil && item.Request.Limit == 30 && !reflect.DeepEqual(25, len(result.Users)) {
			t.Errorf("[%d] Error test case, not len(testCase) %d with len(result) %d", caseNum, len(item.Result.Users), len(result.Users))
		}

		if err == nil && item.Request.Limit == 5 && !reflect.DeepEqual(5, len(result.Users)) {
			t.Errorf("[%d] Error test case, not len(testCase) %d with len(result) %d", caseNum, len(item.Result.Users), len(result.Users))
		}

		if err == nil && item.IsError {
			t.Errorf("[%d] Error: %v, %v", caseNum, result, err)
		}
	}

	///////Error Token!

	caseOne := &TestCase{
		Request: &SearchRequest{
			Limit:  5,
			Offset: 50,
		},
		Result: &SearchResponse{
			Users:    []User{},
			NextPage: false,
		},
		IsError: true,
	}

	ts = httptest.NewServer(http.HandlerFunc(SearchServerErrorToken))

	c := &SearchClient{
		AccessToken: "Kek",
		URL:         ts.URL,
	}

	_, err := c.FindUsers(*caseOne.Request)

	if !(err != nil) && caseOne.IsError {
		t.Errorf("Error AccessToken check")
	}

	///Unpack with bad request code

	caseOne = &TestCase{
		Request: &SearchRequest{
			Limit:  5,
			Offset: 2,
		},
		Result: &SearchResponse{
			Users:    []User{},
			NextPage: false,
		},
		IsError: true,
	}

	ts = httptest.NewServer(http.HandlerFunc(SearchServerErrorUnpackJson))

	c = &SearchClient{
		AccessToken: "Kek",
		URL:         ts.URL,
	}

	_, err = c.FindUsers(*caseOne.Request)

	if !(err != nil) && caseOne.IsError {
		t.Errorf("Error Unpack json check with bad request status code (400)")
	}

	///Unpack with success status code

	caseOne = &TestCase{
		Request: &SearchRequest{
			Limit:  5,
			Offset: 2,
		},
		Result: &SearchResponse{
			Users:    []User{},
			NextPage: false,
		},
		IsError: true,
	}

	ts = httptest.NewServer(http.HandlerFunc(SearchServerErrorOkUnpackJson))

	c = &SearchClient{
		AccessToken: "",
		URL:         ts.URL,
	}

	_, err = c.FindUsers(*caseOne.Request)

	if !(err != nil) && caseOne.IsError {
		t.Errorf("Error Unpack json check with success status code (200)")
	}

	///Unpack with success status code

	caseOne = &TestCase{
		Request: &SearchRequest{
			Limit:  5,
			Offset: 2,
		},
		Result: &SearchResponse{
			Users:    []User{},
			NextPage: false,
		},
		IsError: true,
	}

	ts = httptest.NewServer(http.HandlerFunc(SearchServerErrorInternalServer))

	c = &SearchClient{
		AccessToken: "",
		URL:         ts.URL,
	}

	_, err = c.FindUsers(*caseOne.Request)

	if !(err != nil) && caseOne.IsError {
		t.Errorf("Error SearchServerErrorInternalServer with internal server status code (500)")
	}

	///SearchServerErrorUnknowStatus (code 502)

	caseOne = &TestCase{
		Request: &SearchRequest{
			Limit:  5,
			Offset: 2,
		},
		Result: &SearchResponse{
			Users:    []User{},
			NextPage: false,
		},
		IsError: true,
	}

	ts = httptest.NewServer(http.HandlerFunc(SearchServerErrorUnknowStatus))

	c = &SearchClient{
		AccessToken: "",
		URL:         ts.URL,
	}

	_, err = c.FindUsers(*caseOne.Request)

	if !(err != nil) && caseOne.IsError {
		t.Errorf("Error SearchServerErrorUnknowStatus with bad gateway status code (502)")
	}

	ts = httptest.NewServer(http.HandlerFunc(SearchServerErrorUnknowStatus))

	c = &SearchClient{
		AccessToken: "",
		URL:         ts.URL,
	}

	_, err = c.FindUsers(*caseOne.Request)

	if !(err != nil) && caseOne.IsError {
		t.Errorf("Error SearchServerErrorUnknowStatus with bad gateway status code (502)")
	}

	// Timeout

	ts = httptest.NewServer(http.HandlerFunc(SearchServerErrorTimeout))

	c = &SearchClient{
		AccessToken: "",
		URL:         ts.URL,
	}

	_, err = c.FindUsers(*caseOne.Request)
	fmt.Println(err)
	if !(err != nil) && caseOne.IsError {
		t.Errorf("Error Timeout with request timeout status code (408)")
	}

	////unknow error

	ts = &httptest.Server{}

	c = &SearchClient{
		AccessToken: "",
		URL:         ts.URL,
	}

	_, err = c.FindUsers(*caseOne.Request)
	if !(err != nil) && caseOne.IsError {
		t.Errorf("Error Timeout with request timeout status code (408)")
	}
}
