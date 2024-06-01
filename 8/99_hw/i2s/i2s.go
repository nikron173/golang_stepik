package main

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {
	// todo
	if reflect.ValueOf(out).Kind() != reflect.Pointer {
		return fmt.Errorf("Object not pointer")
	}

	switch reflect.ValueOf(out).Elem().Kind() {

	case reflect.Slice:

		s := reflect.ValueOf(out).Elem()
		kek, ok := data.([]interface{})

		if !ok {
			return fmt.Errorf("Error type")
		}

		objectType := reflect.TypeOf(s.Interface()).Elem()
		// fmt.Println(objectType)
		for _, objectMap := range kek {
			newObjectPtr := reflect.New(objectType)
			i2s(objectMap, newObjectPtr.Interface())
			s.Set(reflect.Append(s, newObjectPtr.Elem()))
		}

	case reflect.Struct:
		{
			kek, _ := data.(map[string]interface{})
			s := reflect.ValueOf(out).Elem()

			for i := range s.NumField() {
				switch s.Type().Field(i).Type.Kind() {

				case reflect.Int:
					{
						val, ok := kek[s.Type().Field(i).Name].(float64)
						if !ok {
							return fmt.Errorf("Error type")
						}
						s.Field(i).SetInt(int64(val))
					}
				case reflect.String:
					{
						val, ok := kek[s.Type().Field(i).Name].(string)
						if !ok {
							return fmt.Errorf("Error type")
						}
						s.Field(i).SetString(val)
					}
				case reflect.Float64:
					{
						val, ok := kek[s.Type().Field(i).Name].(float64)
						if !ok {
							return fmt.Errorf("Error type")
						}
						s.Field(i).SetFloat(float64(val))
					}
				case reflect.Bool:
					{
						val, ok := kek[s.Type().Field(i).Name].(bool)
						if !ok {
							return fmt.Errorf("Error type")
						}
						s.Field(i).SetBool(val)
					}
				case reflect.Slice:
					{
						sliceMap, ok := kek[s.Type().Field(i).Name].([]interface{})
						if !ok {
							return fmt.Errorf("Error type")
						}

						objectType := reflect.TypeOf(s.Field(i).Interface()).Elem()

						for _, objectMap := range sliceMap {
							newObjectPtr := reflect.New(objectType)
							i2s(objectMap, newObjectPtr.Interface())
							s.Field(i).Set(reflect.Append(s.Field(i), newObjectPtr.Elem()))
						}
					}
				case reflect.Array:
					{
						fmt.Println("Array")
					}
				default:
					{
						newObjectPtr := reflect.New(s.Field(i).Type())
						val := kek[s.Type().Field(i).Name]
						i2s(val, newObjectPtr.Interface())
						s.Field(i).Set(newObjectPtr.Elem())
					}
				}
			}
		}
	}
	return nil
}

func main() {
	// expected := &Simple{
	// 	ID:       42,
	// 	Username: "rvasily",
	// 	Active:   true,
	// }

	smpl := Simple{
		ID:       42,
		Username: "rvasily",
		Active:   true,
	}
	// expected := &Complex{
	// 	SubSimple:  smpl,
	// 	ManySimple: []Simple{smpl, smpl},
	// 	Blocks:     []IDBlock{IDBlock{42}, IDBlock{42}},
	// }
	expected := []Simple{smpl, smpl}

	jsonRaw, _ := json.Marshal(expected)
	var tmpData interface{}
	json.Unmarshal(jsonRaw, &tmpData)
	l := []Simple{}
	i2s(tmpData, l)
	fmt.Println(tmpData)

}

// type Simple struct {
// 	ID       int
// 	Username string
// 	Active   bool
// }

// type IDBlock struct {
// 	ID int
// }

// type Complex struct {
// 	SubSimple  Simple
// 	ManySimple []Simple
// 	Blocks     []IDBlock
// }
