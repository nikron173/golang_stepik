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
		// fmt.Printf("%s: %v\n", s.Type().Field(i).Name, kek[s.Type().Field(i).Name])
		switch s.Type().Field(i).Type.String() {
		case reflect.Int.String():
			val, ok := kek[s.Type().Field(i).Name].(float64)
			if !ok {
				return fmt.Errorf("Error type")
			}
			s.Field(i).SetInt(int64(val))
		case reflect.String.String():
			val, ok := kek[s.Type().Field(i).Name].(string)
			if !ok {
				return fmt.Errorf("Error type")
			}
			s.Field(i).SetString(val)
		case reflect.Float64.String():
			val, ok := kek[s.Type().Field(i).Name].(float64)
			if !ok {
				return fmt.Errorf("Error type")
			}
			s.Field(i).SetFloat(float64(val))
		case reflect.Bool.String():
			val, ok := kek[s.Type().Field(i).Name].(bool)
			if !ok {
				return fmt.Errorf("Error type")
			}
			s.Field(i).SetBool(val)
		}
	}
	// fmt.Println(out)

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
	fmt.Println(tmpData)
}
