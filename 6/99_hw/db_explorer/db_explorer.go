package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные
type ApplicationError struct {
	HttpCode     int
	ErrorMessage string
}

type Param struct {
	Limit  int
	Offset int
}

func (e ApplicationError) Error() string {
	return fmt.Sprintf("error: %s", e.ErrorMessage)
}

type Handler struct {
	DB *sql.DB
}

type Router struct {
	route []string
}

type Table struct {
	Name      string
	Variables map[string]InfoTable
}

type InfoTable struct {
	Type      string
	NullField bool
	Key       string
}

type Data struct {
	Value []interface{}
}

// func (h *Handler) ListTable(w http.ResponseWriter, r *http.Request) {
// 	tables, err := Info(h.DB)
// 	if err != nil {
// 		appErr, ok := err.(ApplicationError)
// 		if ok {
// 			w.WriteHeader(appErr.HttpCode)
// 			fmt.Fprintf(w, "%s", appErr.ErrorMessage)
// 			return
// 		}
// 		w.WriteHeader(r.Response.StatusCode)
// 		fmt.Fprintf(w, "%s", err)
// 		return
// 	}
// 	w.WriteHeader(http.StatusOK)
// 	fmt.Fprintf(w, "%s", tables)
// }

func Routes(db *sql.DB) (*Router, error) {
	routes := new(Router)
	result, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, ApplicationError{
			HttpCode:     http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		}
	}
	for result.Next() {
		table := ""
		result.Scan(&table)
		routes.route = append(routes.route, table)
	}
	return routes, nil
}

func (d *Data) Get(db *sql.DB, table string, id string) (*Data, error) {
	result, err := db.Query("SHOW FULL COLUMNS FROM `?`", table)
	if err != nil {
		return nil, ApplicationError{
			HttpCode:     http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		}
	}
	t := new(Table)
	for result.Next() {
		result.Scan(&t.Name, &t.Variables)
	}
	fmt.Printf("%#v", t)
	return nil, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "application/json")
	fmt.Printf("path: %s\n", r.URL.Path)

	routes, err := Routes(h.DB)
	if err != nil {
		appErr, ok := err.(ApplicationError)
		if ok {
			w.WriteHeader(appErr.HttpCode)
			fmt.Fprintf(w, "%s", appErr.ErrorMessage)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
		return
	}

	if r.URL.Path == "/" {
		w.WriteHeader(http.StatusOK)
		routesJson, _ := json.Marshal(routes.route)
		fmt.Fprintf(w, "%s", routesJson)
		return
	}

	for _, route := range routes.route {
		if route == strings.Split(r.URL.Path, "/")[1] {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Table %s found", route)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "Page %s not found", r.URL.Path)
}

func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	h := &Handler{
		DB: db,
	}
	return h, nil
}
