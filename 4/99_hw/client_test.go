package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type TestCase struct {
	Id      string
	Result  *CheckoutResult
	IsError bool
}

type CheckoutResult struct {
	Status  int
	Balance int
	Err     string
}

// код писать тут
func SearchServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "application/json")
	if r.Method != http.MethodGet {
		http.ErrNotSupported.Error()
		return
	}
	param := new(SearchRequest)
	param.OrderField = "Name"
	param.Limit = 10

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
				Error: "Error parameter order field",
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

	response := &SearchResponse{}

	if param.Offset >= len(findUsers) {
		findUsers = make([]User, 0)
		w.WriteHeader(http.StatusOK)
		response.Users = findUsers
		jsonResponse, _ := json.Marshal(response)
		fmt.Fprintf(w, "%s", jsonResponse)
		return
	}

	if param.Limit > 0 && param.Limit <= len(findUsers) && (param.Offset+1) > 0 && (param.Offset+1) <= len(findUsers) {
		if param.Limit+param.Offset > len(findUsers) {
			findUsers = findUsers[param.Offset:]
		} else {
			findUsers = findUsers[param.Offset : param.Offset+param.Limit]
			response.NextPage = true
		}
	}

	response.Users = findUsers
	w.WriteHeader(http.StatusOK)
	jsonResponse, _ := json.Marshal(response)
	fmt.Fprintf(w, "%s", jsonResponse)
}

func TestFindUser(t *testing.T) {
	cases := []TestCase{}
}
