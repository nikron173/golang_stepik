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

func (d *Data) Get(db *sql.DB, table string, id string) (*Data, error) {
	// q := fmt.Sprintf("SHOW FULL COLUMNS FROM ?", table)
	// result, err := db.Query(q)
	result, err := db.Query("SELECT * FROM ?", table)
	if err != nil {
		fmt.Printf(err.Error())
		return nil, ApplicationError{
			HttpCode:     http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		}
	}

	t := new(Table)
	t.Name = table
	fmt.Printf("%#v\n", t)
	r, _ := result.Columns()
	tt, _ := result.ColumnTypes()
	fmt.Printf("Columns: %s\n", r)
	fmt.Printf("Types columns: %s\n", tt)

	for result.Next() {
		info := new(InfoTable)
		result.Scan(&info.Field, &info.Key, &info.Type, &info.NullField)
		// fmt.Printf("%#v", info)
	}
	fmt.Printf("%#v", t)
	return nil, nil
}

func main() {
	db, err := sql.Open("mysql", "root:love@tcp(localhost:3306)/photolist?charset=utf8")
	err = db.Ping() // вот тут будет первое подключение к базе
	if err != nil {
		panic(err)
	}

	fmt.Printf("Successfull connected database:\n %#v\n", db.Stats())

	d := new(Data)
	d.Get(db, "items", "1")
}
