package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// код писать тут

var (
	// type CreateParams struct {
	// 	Login  string `apivalidator:"required,min=10"`
	// 	Name   string `apivalidator:"paramname=full_name"`
	// 	Status string `apivalidator:"enum=user|moderator|admin,default=user"`
	// 	Age    int    `apivalidator:"min=0,max=128"`
	// }

	apivalidatorFieldReg = regexp.MustCompile(`(\")(.*)(\")`)
)

type Apigen struct {
	Url                  string
	Auth                 bool
	Method               string
	StructureHandlerName string
	FunctionApiName      string
	StructureInParam     []ApigenParam
}

type ApigenParam struct {
	ParamName string
	ParamType string
}

// type ApivalidatorStruct struct {
// 	StructName string
// 	Variable   []ApivalidatorField
// }

type ApivalidatorField struct {
	VariableName string
	VariableType string
	Rq           Requirement
}

type Requirement struct {
	Required  bool
	Paramname string
	Enum      []string
	Min       string
	Max       string
	Default   string
}

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := os.Create(os.Args[2])

	fmt.Fprintln(out, `package `+node.Name.Name)
	fmt.Fprintln(out) // empty line

	fmt.Fprintln(out, `import (`)
	fmt.Fprintln(out, `		"encoding/json"`)
	fmt.Fprintln(out, `		"fmt"`)
	fmt.Fprintln(out, `		"net/http"`)
	fmt.Fprintln(out, `		"strconv"`)
	fmt.Fprintln(out, `)`)

	fmt.Fprintln(out) // empty line
	// fmt.Fprintln(out, "type CR map[string]interface{}") // empty line
	fmt.Fprintln(out) // empty line

	apiGenMap, err := generateApiGenArr(node)
	if err != nil {
		panic(err)
	}

	apivalidatorFields, err := generateApivalidatorFields(node, apiGenMap)
	if err != nil {
		panic(err)
	}

	for k, v := range apiGenMap {
		for _, api := range v {
			fmt.Fprintf(out, "func (h *%s)", k)
			fmt.Fprintf(out, " handler%s%s(w http.ResponseWriter, r *http.Request) {\n", k, api.FunctionApiName)

			if strings.ToUpper(api.Method) == http.MethodPost {
				fmt.Fprintln(out, `
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotAcceptable)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}", "error", "bad method")
		return
	}
`)
			}

			if api.Auth {
				fmt.Fprintln(out, `
	if r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}", "error", "unauthorized")
		return
	}
				`)
			}

			// 			fmt.Fprintf(out, `
			// 	if err := r.ParseForm(); err != nil {
			// 		w.WriteHeader(http.StatusBadRequest)
			// 		fmt.Fprintf(w, "{\"%s\": \"%s\"}",
			// 			"error", err.Error())
			// 		return
			// 	}
			// `)

			for _, localVar := range api.StructureInParam {
				fmt.Fprintf(out, "\t%s := new(%s)\n", localVar.ParamName, localVar.ParamType)
				validatorField, ok := apivalidatorFields[localVar.ParamType]
				if !ok {
					continue
				}
				for _, parameterCheck := range validatorField {
					if parameterCheck.Rq.Paramname != "" {
						fmt.Fprintf(out, `
	%sStr := r.FormValue("%s")

`, strings.ToLower(parameterCheck.VariableName), parameterCheck.Rq.Paramname)
					} else {
						fmt.Fprintf(out, `
	%sStr := r.FormValue("%s")
	
`, strings.ToLower(parameterCheck.VariableName), strings.ToLower(parameterCheck.VariableName))
					}

					if parameterCheck.Rq.Required {
						fmt.Fprintf(out, "\tif %sStr == \"\" {", strings.ToLower(parameterCheck.VariableName))
						fmt.Fprintln(out, `
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}",
`)
						fmt.Fprintf(out, `
			"error", "%s must me not empty")
		return
	}
`, strings.ToLower(parameterCheck.VariableName))
					}

					if parameterCheck.VariableType == "int" {
						fmt.Fprintf(out, "\t%s, err := strconv.Atoi(%sStr)\n", strings.ToLower(parameterCheck.VariableName), strings.ToLower(parameterCheck.VariableName))
						fmt.Fprintln(out, `
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}",`)
						fmt.Fprintf(out, `
			"error", "%s must be int")
		return
	}
`, strings.ToLower(parameterCheck.VariableName))
					} else {
						fmt.Fprintf(out, `
	%s := %sStr
`, strings.ToLower(parameterCheck.VariableName), strings.ToLower(parameterCheck.VariableName))
					}

					if parameterCheck.Rq.Min != "" {
						if parameterCheck.VariableType == "int" {
							fmt.Fprintf(out, "\tif %s < %s {", strings.ToLower(parameterCheck.VariableName), parameterCheck.Rq.Min)
							fmt.Fprintln(out, `
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}",`)
							fmt.Fprintf(out, `
			"error", "%s must be >= %s")
		return
	}
`, strings.ToLower(parameterCheck.VariableName), parameterCheck.Rq.Min)
						} else {
							fmt.Fprintf(out, "\tif len(%s) < %s {", strings.ToLower(parameterCheck.VariableName), parameterCheck.Rq.Min)
							fmt.Fprintln(out, `
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}",`)
							fmt.Fprintf(out, `
			"error", "%s len must be >= %s")
		return
	}
`, strings.ToLower(parameterCheck.VariableName), parameterCheck.Rq.Min)
						}
					}

					if parameterCheck.Rq.Max != "" {
						if parameterCheck.VariableType == "int" {
							fmt.Fprintf(out, "\tif %s > %s {", strings.ToLower(parameterCheck.VariableName), parameterCheck.Rq.Max)
							fmt.Fprintln(out, `
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}",`)
							fmt.Fprintf(out, `
			"error", "%s must be <= %s")
		return
	}
`, strings.ToLower(parameterCheck.VariableName), parameterCheck.Rq.Max)
						} else {
							fmt.Fprintf(out, "\tif len(%s) > %s {", strings.ToLower(parameterCheck.VariableName), parameterCheck.Rq.Max)
							fmt.Fprintln(out, `
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}",`)
							fmt.Fprintf(out, `
			"error", "%s must be <= %s")
		return
	}
`, strings.ToLower(parameterCheck.VariableName), parameterCheck.Rq.Max)
						}
					}

					if parameterCheck.Rq.Default != "" && parameterCheck.VariableType == "int" {
						fmt.Fprintf(out, `
	if %s == 0 {
		%s = %s
	}
`, strings.ToLower(parameterCheck.VariableName), strings.ToLower(parameterCheck.VariableName), parameterCheck.Rq.Default)
					} else if parameterCheck.Rq.Default != "" && parameterCheck.VariableType == "string" {
						fmt.Fprintf(out, `
	if %s == "" {
		%s = "%s"
	}
`, strings.ToLower(parameterCheck.VariableName), strings.ToLower(parameterCheck.VariableName), parameterCheck.Rq.Default)
					}

					if len(parameterCheck.Rq.Enum) != 0 {
						fmt.Fprintf(out, "\tif !(")
						if parameterCheck.VariableType == "int" {
							for index, p := range parameterCheck.Rq.Enum {
								if index == len(parameterCheck.Rq.Enum)-1 {
									fmt.Fprintf(out, "%s == %s) {", p, strings.ToLower(parameterCheck.VariableName))
									break
								}
								fmt.Fprintf(out, "%s == %s || ", p, strings.ToLower(parameterCheck.VariableName))
							}
						} else {
							for index, p := range parameterCheck.Rq.Enum {
								if index == len(parameterCheck.Rq.Enum)-1 {
									fmt.Fprintf(out, "\"%s\" == %s) {", p, strings.ToLower(parameterCheck.VariableName))
									break
								}
								fmt.Fprintf(out, "\"%s\" == %s || ", p, strings.ToLower(parameterCheck.VariableName))
							}
						}

						fmt.Fprintln(out, `
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}",`)
						fmt.Fprintf(out, `
			"error", "%s must be one of [%s]")
		return
	}
`, strings.ToLower(parameterCheck.VariableName), strings.Join(parameterCheck.Rq.Enum, ", "))
					}
					fmt.Fprintf(out, "\t%s.%s = %s\n",
						localVar.ParamName, parameterCheck.VariableName, strings.ToLower(parameterCheck.VariableName))
				}
			}

			fmt.Fprintf(out, "\tres, err := h.%s(r.Context(), *%s)", api.FunctionApiName, string(api.StructureInParam[0].ParamName))
			fmt.Fprintln(out, `
	if err != nil {
		errApiError, ok := err.(ApiError)
		if ok {
			w.WriteHeader(errApiError.HTTPStatus)
			fmt.Fprintf(w, "{\"%s\": \"%s\"}",
			"error", errApiError)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}",
			"error", err)
		return
	}
	cr := CR{
		"error": "",
		"response": res,
	}

	w.WriteHeader(http.StatusOK)
	resCr, _ := json.Marshal(cr)
	fmt.Fprintf(w, "%s", resCr)
}
`)
		}
	}

	for k, v := range apiGenMap {
		fmt.Fprintf(out, "func (h *%s) ServeHTTP(w http.ResponseWriter, r *http.Request) {\n", k)
		fmt.Fprintln(out, `	w.Header().Add("Content-type", "application/json")
	switch r.URL.Path {`)
		for _, api := range v {
			fmt.Fprintf(out, "	case \"%s\":\n", api.Url)
			fmt.Fprintf(out, "		h.handler%s%s(w, r)\n", k, api.FunctionApiName)
		}
		fmt.Fprintln(out, `	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}",
			"error", "unknown method")
	}
}
		`)
	}

	// fmt.Println(apiGenMap)
	// fmt.Println(apivalidatorFields)
}

func generateApiGenArr(node *ast.File) (map[string][]Apigen, error) {
	apiGenArr := make([]Apigen, 0)
	apiGenMap := make(map[string][]Apigen)

	count := 0
	for _, d := range node.Decls {
		f, ok := d.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if f.Doc == nil {
			continue
		}
		needgen := false
		for _, com := range f.Doc.List {
			if com != nil && strings.Contains(com.Text, "apigen") {
				needgen = true
				apigen := new(Apigen)
				apigenStr, _ := strings.CutPrefix(com.Text, "// apigen:api ")
				if err := json.Unmarshal([]byte(apigenStr), apigen); err != nil {
					return nil, fmt.Errorf("Ошибка парбсинга комментария в структуру у метода: %s, comment: '%s'",
						f.Name.Name, com.Text)
				}
				apiGenArr = append(apiGenArr, *apigen)
			}
		}

		if needgen {
			apiGenArr[count].FunctionApiName = f.Name.Name

			// разворачивает то, что имплементируется (структура)
			if f.Recv != nil {
				for _, ff := range f.Recv.List {
					if recv, ok := ff.Type.(*ast.StarExpr); ok {
						if recvIdent, ok := recv.X.(*ast.Ident); ok {
							apiGenArr[count].StructureHandlerName = recvIdent.Name
						}
					}
				}
			}

			for _, ff := range f.Type.Params.List {
				apigenParam := make([]ApigenParam, 0)

				for _, paramName := range ff.Names {
					if tt, ok := ff.Type.(*ast.Ident); ok {
						ap := new(ApigenParam)
						ap.ParamName = paramName.Name
						ap.ParamType = tt.Name
						apigenParam = append(apigenParam, *ap)
					}
				}
				apiGenArr[count].StructureInParam = append(apiGenArr[count].StructureInParam, apigenParam...)
			}
			count++
		}
	}
	for _, a := range apiGenArr {
		apiGenMap[a.StructureHandlerName] = append(apiGenMap[a.StructureHandlerName], a)
	}
	return apiGenMap, nil
}

func generateApivalidatorFields(node *ast.File, apiGenMap map[string][]Apigen) (map[string][]ApivalidatorField, error) {
	aSArray := make(map[string][]ApivalidatorField)
	for _, f := range node.Decls {
		g, ok := f.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range g.Specs {

			currType, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			currStruct, ok := currType.Type.(*ast.StructType)
			if !ok {
				continue
			}

			needCodegenStruct := false
		LOOP_CHECH:
			for _, apiGenArr := range apiGenMap {
				for _, strt := range apiGenArr {
					for _, strtName := range strt.StructureInParam {
						if currType.Name.Name == strtName.ParamType {
							needCodegenStruct = true
							break LOOP_CHECH
						}
					}
				}
			}

			if !needCodegenStruct {
				continue
			}
			fmt.Printf("Current Sturct: %v\n", currType.Name.Name)
			// aS.StructName = currType.Name.Name
			aSArray[currType.Name.Name] = make([]ApivalidatorField, 0)

			for _, field := range currStruct.Fields.List {
				if field.Tag == nil {
					continue
				}
				aF := new(ApivalidatorField)

				if ok := strings.Contains(field.Tag.Value, "apivalidator"); ok {
					tt, _ := field.Type.(*ast.Ident)
					fmt.Printf("Name and type variable: %s and %s\n", field.Names[0].Name, tt.Name)
					aF.VariableName = field.Names[0].Name
					aF.VariableType = tt.Name

					//fmt.Printf("%s: %s\n", "apivalidatorField.FindString(field.Tag.Value)", apivalidatorField.FindString(field.Tag.Value))
					for _, apiField := range strings.Split(apivalidatorFieldReg.FindString(field.Tag.Value), ",") {
						apiField = strings.Replace(apiField, "\"", "", -1)
						// fmt.Printf("%s: %s\n", "ApiField", apiField)
						if strings.Contains(apiField, "required") {
							fmt.Printf("%s: %s\n", "надо добавить обязательное поле required", strings.ToLower(field.Names[0].Name))
							aF.Rq.Required = true
							continue
						}
						if strings.Contains(apiField, "paramname") {
							reqParam := strings.Split(apiField, "=")[1]
							fmt.Printf("%s: %s\n", "по какому полю искать совпадение paramname", reqParam)
							aF.Rq.Paramname = reqParam
							continue
						}
						if strings.Contains(apiField, "enum") {
							reqEnum := strings.Split(strings.Split(apiField, "=")[1], "|")
							fmt.Printf("%s: %s\n", "значения enum", reqEnum)
							aF.Rq.Enum = reqEnum
							continue
						}
						if strings.Contains(apiField, "min=") {
							minValue := strings.Split(apiField, "=")[1]
							_, err := strconv.Atoi(minValue)
							if err != nil {
								panic(err)
							}
							fmt.Printf("%s: %s\n", "минимальное значение по полю min", minValue)
							aF.Rq.Min = minValue
							continue
						}
						if strings.Contains(apiField, "max=") {
							maxValue := strings.Split(apiField, "=")[1]
							_, err := strconv.Atoi(maxValue)
							if err != nil {
								panic(err)
							}
							fmt.Printf("%s: %s\n", "минимальное значение по полю max", maxValue)
							aF.Rq.Max = maxValue
							continue
						}
						if strings.Contains(apiField, "default") {
							reqDefault := strings.Split(apiField, "=")[1]
							fmt.Printf("%s: %s\n", "дефолтное значение поля, если не указано значение default", reqDefault)
							aF.Rq.Default = reqDefault
							continue
						}
					}
				}
				// aS.Variable = append(aS.Variable, *aF)
				aSArray[currType.Name.Name] = append(aSArray[currType.Name.Name], *aF)
			}
			// aSArray = append(aSArray, *aS)
		}
		fmt.Println()
	}
	return aSArray, nil
}
