package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type User struct {
	Id        int    `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       int    `xml:"age"`
	About     string `xml:"about" json:"-"`
}

type Param struct {
	Query      string
	OrderField string
	OrderBy    string
	Limit      int
	Offset     int
}

func (u User) String() string {
	return fmt.Sprintf("\t{\"id\": %d, \"firstName\": %s, \"lastName\": %s, \"age\": %d},\n",
		u.Id, u.FirstName, u.LastName, u.Age)
}

func SortSliceUserField(u *[]User, param string, orderBy string, limit int, offset int) error {
	switch param {
	case "Id":
		{
			sort.Slice((*u), func(i, j int) bool {
				return (*u)[i].Id > (*u)[j].Id
			})
		}
	case "Age":
		{
			sort.Slice((*u), func(i, j int) bool {
				return (*u)[i].Age > (*u)[j].Age
			})
		}
	case "Name":
		{
			sort.Slice((*u), func(i, j int) bool {
				fullNameOne := (*u)[i].FirstName + " " + (*u)[i].LastName
				fullNameSecond := (*u)[j].FirstName + " " + (*u)[j].LastName
				return fullNameOne > fullNameSecond
			})
		}
	default:
		{
			return &ApplicationError{
				Code:    http.StatusBadRequest,
				Message: "Error sort by field (Id, Name, Age)",
				Time:    time.Now(),
			}
		}
	}

	switch orderBy {
	case "desc":
		{

		}
	case "asc":
		{
			rv := make([]User, len((*u)), len((*u)))
			for i := len((*u)) - 1; i >= 0; i-- {
				rv[len((*u))-i-1] = (*u)[i]
			}
			*u = rv
		}
	default:
		{
			return &ApplicationError{
				Code:    http.StatusBadRequest,
				Message: "Not valid order by parameter (asc/desc)",
				Time:    time.Now(),
			}
		}
	}
	//по умолчанию - это убывание
	if offset >= len(*u) {
		*u = make([]User, 0)
		return nil
	}

	if limit > 0 && limit <= len(*u) && (offset+1) > 0 && (offset+1) <= len(*u) {
		if limit+offset > len(*u) {
			// limitUsers := make([]User, len(*u)-limit, len(*u)-limit)
			*u = (*u)[offset:len(*u)]
		} else {
			// limitUsers := make([]User, limit, limit)
			*u = (*u)[offset : offset+limit]
		}
	}

	return nil
}

type Row struct {
	ListUsers []User `xml:"row"`
}

type ApplicationError struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

func (e *ApplicationError) Error() string {
	return fmt.Sprintf("code: %d, error: %s, time: %s",
		e.Code, e.Message, e.Time)
}

func searchUser(param *Param) ([]User, error) {
	row := &Row{}

	file, err := os.Open("dataset.xml")
	if err != nil {
		return nil, &ApplicationError{
			Code:    http.StatusInternalServerError,
			Message: "Error internal server",
			Time:    time.Now(),
		}
	}

	defer file.Close()
	fileContents, err := io.ReadAll(file)
	if err != nil {
		return nil, &ApplicationError{
			Code:    http.StatusInternalServerError,
			Message: "Error internal server",
			Time:    time.Now(),
		}
	}

	if err := xml.Unmarshal(fileContents, row); err != nil {
		return nil, &ApplicationError{
			Code:    http.StatusInternalServerError,
			Message: "Error parsing xml to golang objects",
			Time:    time.Now(),
		}
	}

	findUsers := make([]User, 0)

	if len(param.Query) == 0 {
		findUsers = append(findUsers, row.ListUsers...)
	} else {
		for _, user := range row.ListUsers {
			if strings.Contains(user.About, param.Query) ||
				strings.Contains(user.FirstName, param.Query) ||
				strings.Contains(user.LastName, param.Query) {
				findUsers = append(findUsers, user)
			}
		}
	}

	if err := SortSliceUserField(&findUsers, param.OrderField, param.OrderBy, param.Limit, param.Offset); err != nil {
		return nil, err
	}
	return findUsers, nil
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "application/json")
	if r.Method != http.MethodGet {
		http.ErrNotSupported.Error()
		return
	}
	param := new(Param)
	param.OrderField = "Name"
	param.OrderBy = "desc"

	var err error

	if query := r.FormValue("query"); query != "" {
		param.Query = query
	}
	if orderField := r.FormValue("order_field"); orderField != "" {
		param.OrderField = orderField
	}
	if orderBy := r.FormValue("order_by"); orderBy != "" {
		param.OrderBy = orderBy
	}
	if limitStr := r.FormValue("limit"); limitStr != "" {
		if param.Limit, err = strconv.Atoi(limitStr); err != nil || param.Limit < 0 {
			appErr := &ApplicationError{
				Code:    http.StatusBadRequest,
				Message: "Error parameter limit",
				Time:    time.Now(),
			}
			appErrJson, _ := json.Marshal(appErr)
			w.WriteHeader(appErr.Code)
			fmt.Fprintf(w, "%s", appErrJson)
			return
		}
	}
	if offsetStr := r.FormValue("offset"); offsetStr != "" {
		if param.Offset, err = strconv.Atoi(offsetStr); err != nil || param.Offset < 0 {
			appErr := &ApplicationError{
				Code:    http.StatusBadRequest,
				Message: "Error parameter offset",
				Time:    time.Now(),
			}
			appErrJson, _ := json.Marshal(appErr)
			w.WriteHeader(appErr.Code)
			fmt.Fprintf(w, "%s", appErrJson)
			return
		}
	}

	users, err := searchUser(param)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			w.WriteHeader(appErr.Code)
		} else {
			w.WriteHeader(http.StatusBadGateway)
		}
		errJson, _ := json.Marshal(err)
		fmt.Fprintf(w, "%s", errJson)
		return
	}

	jsonUsers, err := json.Marshal(users)
	if err != nil {
		if appErr, ok := err.(*ApplicationError); ok {
			w.WriteHeader(appErr.Code)
		} else {
			w.WriteHeader(http.StatusBadGateway)
		}
		errJson, _ := json.Marshal(err)
		fmt.Fprintf(w, "%s", errJson)
		return
	}

	fmt.Fprintf(w, "%s", jsonUsers)

}

func main() {

	// r := bufio.NewReader(os.Stdin)
	// fmt.Println("Enter user name:")
	// name, _ := r.ReadString('\n')
	// name = strings.Replace(name, "\r\n", "", -1)
	// fmt.Println("Enter user about:")
	// about, _ := r.ReadString('\n')
	// about = strings.Replace(about, "\r\n", "", -1)
	// fmt.Println("Enter user order field:")
	// orderField, _ := r.ReadString('\n')
	// orderField = strings.Replace(orderField, "\r\n", "", -1)
	// fmt.Println("Enter user order by (asc/desc):")
	// orderBy, _ := r.ReadString('\n')
	// orderBy = strings.Replace(orderBy, "\r\n", "", -1)
	// fmt.Println("Enter user limit (int):")
	// limitStr, _ := r.ReadString('\n')
	// limit, err := strconv.Atoi(strings.Replace(limitStr, "\r\n", "", -1))
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("Enter user offset (int):")
	// offsetStr, _ := r.ReadString('\n')
	// offset, err := strconv.Atoi(strings.Replace(offsetStr, "\r\n", "", -1))
	// if err != nil {
	// 	panic(err)
	// }

	// param := &Param{}
	// param.Name = name
	// param.About = about
	// param.OrderField = orderField
	// param.OrderBy = orderBy
	// param.Limit = limit
	// param.Offset = offset

	// fmt.Println(SearchUser(param))

	http.HandleFunc("/users", SearchServer)

	fmt.Println("Starting web server on port :8081")
	http.ListenAndServe(":8081", nil)
}
