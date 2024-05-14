package main

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные
type ApplicationError struct {
	HttpCode     int
	ErrorMessage string
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
	Field     string
	Type      string
	NullField bool
	Key       string
}

type Data struct {
	Value []interface{}
}

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

type Info interface{}

func (h *Handler) GetObject(table string, id string) (*Data, error) {
	q := fmt.Sprintf("SHOW FULL COLUMNS FROM `%s`", table)
	result, err := h.DB.Query(q)
	if err != nil {
		fmt.Printf(err.Error())
		return nil, ApplicationError{
			HttpCode:     http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		}
	}

	r, _ := result.Columns()

	tt, _ := result.ColumnTypes()
	fmt.Printf("Columns: %s\n\n", r)
	fmt.Printf("Types columns: %s\n\n", tt)

	// kek := make([]interface{}, len(r))
	kek := make([]Info, len(r))
	for result.Next() {

		// result.Scan(&kek[0], &kek[1], &kek[2], &kek[3], &kek[4], &kek[5], &kek[6], &kek[7], &kek[8])
		result.Scan(&kek)
	}
	fmt.Printf("%#v", kek)
	return nil, nil
}

func main() {
	db, err := sql.Open("mysql", "root:love@tcp(localhost:3306)/photolist?charset=utf8")
	err = db.Ping() // вот тут будет первое подключение к базе
	if err != nil {
		panic(err)
	}

	fmt.Printf("Successfull connected database:\n %#v\n", db.Stats())

	h := &Handler{
		DB: db,
	}

	h.GetObject("items", "1")
}
