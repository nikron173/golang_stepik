package main

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {
	// todo
	s := reflect.ValueOf(out).Elem()
	// typeOfS := s.Type()
	kek, _ := data.(map[string]interface{})
	for i := range s.NumField() {
		// j := s.Field(i)
		// fmt.Printf("%d: %s, %s, %v\n", i, s.Type().Field(i).Name, j.Type(), j.Interface())
		fmt.Printf("%s: %v\n", s.Type().Field(i).Name, kek[s.Type().Field(i).Name])
		switch s.Type().Field(i).Type.String() {
		case reflect.Int.String():
			val := (int64)(kek[s.Type().Field(i).Name].(float64))
			s.Field(i).SetInt(val)
		}
		// s.Field(i).Set(reflect.ValueOf(kek[s.Type().Field(i).Name]))
	}
	fmt.Println(out)

	return nil
}

type Simple struct {
	ID       int
	Username string
	Active   bool
}

func main() {
	expected := &Simple{
		ID:       42,
		Username: "rvasily",
		Active:   true,
	}
	jsonRaw, _ := json.Marshal(expected)
	var tmpData interface{}
	json.Unmarshal(jsonRaw, &tmpData)

	i2s(tmpData, new(Simple))
}
