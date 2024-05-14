package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные
type Handler struct {
	DB    *sql.DB
	Table map[string][]InfoTable
}

// type Table struct {
// 	Name string
// 	Info InfoTable
// }

type InfoTable struct {
	Field      interface{}
	Type       interface{}
	Collation  interface{}
	Null       interface{}
	Key        interface{}
	Default    interface{}
	Extra      interface{}
	Privileges interface{}
	Comment    interface{}
}

type Data struct {
	Rows []Row
}

type Row []interface{}

type FrontData struct {
	Key   string
	Value interface{}
}

type Response map[string]interface{}

type Param struct {
	Limit  int
	Offset int
}

func (h *Handler) CreateQueryGetById(tableName string) string {
	var q strings.Builder
	q.WriteString("SELECT ")

	for index, row := range h.Table[tableName] {
		if index == len(h.Table[tableName])-1 {
			q.WriteString(fmt.Sprintf("%s ", row.Field))
			continue
		}
		q.WriteString(fmt.Sprintf("%s", row.Field) + ", ")
	}
	q.WriteString(fmt.Sprintf("FROM %s WHERE ", tableName))
	for _, row := range h.Table[tableName] {
		key := fmt.Sprintf("%s", row.Key)
		if key == "PRI" {
			q.WriteString(fmt.Sprintf("%s = ?", row.Field))
		}
	}
	fmt.Println(q.String())
	return q.String()
}

func (h *Handler) CreateQueryGetRecords(tableName string, param *Param) string {
	var q strings.Builder
	q.WriteString("SELECT ")

	for index, row := range h.Table[tableName] {
		if index == len(h.Table[tableName])-1 {
			q.WriteString(fmt.Sprintf("%s ", row.Field))
			continue
		}
		q.WriteString(fmt.Sprintf("%s", row.Field) + ", ")
	}
	q.WriteString(fmt.Sprintf("FROM %s", tableName))
	// if param.Limit != -1 {
	// 	q.WriteString(fmt.Sprintf(" LIMIT %d", param.Limit))
	// }
	// if param.Offset != -1 {
	// 	q.WriteString(fmt.Sprintf(" OFFSET %d", param.Offset))
	// }
	q.WriteString(fmt.Sprintf(" LIMIT %d", param.Limit))
	q.WriteString(fmt.Sprintf(" OFFSET %d", param.Offset))
	fmt.Println(q.String())
	return q.String()
}

func (h *Handler) CreateQueryDeleteById(tableName string) string {
	var q strings.Builder
	q.WriteString("DELETE FROM ")
	q.WriteString(fmt.Sprintf("%s WHERE ", tableName))
	for _, row := range h.Table[tableName] {
		key := fmt.Sprintf("%s", row.Key)
		if key == "PRI" {
			q.WriteString(fmt.Sprintf("%s = ?", row.Field))
		}
	}
	fmt.Println(q.String())
	return q.String()
}

func (h *Handler) StorageDeleteById(tableName string, reqId string) (map[string]interface{}, error) {
	q := h.CreateQueryDeleteById(tableName)
	result, err := h.DB.Exec(q, reqId)

	if err != nil {
		return nil, &ApplicationError{
			HttpCode:     http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		}
	}
	res, err := result.RowsAffected()
	if err != nil {
		return nil, &ApplicationError{
			HttpCode:     http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		}
	}

	deleted := make(map[string]interface{})
	deleted["deleted"] = res
	return deleted, nil
}

func (h *Handler) CreateQueryInsertById(tableName string, fData map[string]FrontData) (string, Row, error) {
	var q strings.Builder
	var qEnd strings.Builder
	q.WriteString("INSERT INTO ")
	q.WriteString(fmt.Sprintf("%s (", tableName))
	qEnd.WriteString(" VALUES (")
	zapytay := false
	rowValue := make(Row, 0)
	insert := false
	for index, row := range h.Table[tableName] {
		insert = false
		field := fmt.Sprintf("%s", row.Field)
		key := fmt.Sprintf("%s", row.Key)
		if key == "PRI" {
			continue
		}
		val, ok := fData[field]
		if fmt.Sprintf("%s", row.Null) == "YES" && !ok {
			if index == len(h.Table[tableName])-1 {
				q.WriteString(")")
				qEnd.WriteString(")")
				continue
			}
			continue
		}
		if zapytay {
			q.WriteString(", ")
			qEnd.WriteString(", ")
		}

		tField := fmt.Sprintf("%s", row.Type)
		if strings.Contains(tField, "int") {
			valInt, err := strconv.Atoi(fmt.Sprintf("%s", val.Value))
			if err != nil {
				return "", nil, &ApplicationError{
					HttpCode:     http.StatusBadRequest,
					ErrorMessage: fmt.Sprintf("field %s have invalid type", field),
				}
			}
			rowValue = append(rowValue, valInt)
			insert = true
		} else if strings.Contains(tField, "float") || strings.Contains(tField, "double") {
			valFloat, ok := val.Value.(float64)
			if !ok {
				return "", nil, &ApplicationError{
					HttpCode:     http.StatusBadRequest,
					ErrorMessage: fmt.Sprintf("field %s have invalid type", field),
				}
			}
			rowValue = append(rowValue, valFloat)
			insert = true
		}

		if index == len(h.Table[tableName])-1 {
			q.WriteString(fmt.Sprintf("%s)", row.Field))
			qEnd.WriteString("?)")
			if !insert {
				rowValue = append(rowValue, val.Value)
			}
			continue
		}
		if fmt.Sprintf("%s", row.Null) == "NO" && !ok {
			if !insert {
				rowValue = append(rowValue, "")
			}
		} else {
			if !insert {
				rowValue = append(rowValue, val.Value)
			}
		}
		q.WriteString(fmt.Sprintf("%s", row.Field))
		qEnd.WriteString("?")
		zapytay = true
	}

	q.WriteString(qEnd.String())
	return q.String(), rowValue, nil
}

func (h *Handler) StorageInsertObject(tableName string, fData map[string]FrontData) (map[string]interface{}, error) {
	q, rowVal, err := h.CreateQueryInsertById(tableName, fData)
	fmt.Printf("Query: %s\n", q)
	fmt.Printf("rowValue: %v\n", rowVal)
	fmt.Printf("Error: %v\n", err)
	if err != nil {
		return nil, err
	}

	result, err := h.DB.Exec(q, rowVal...)
	if err != nil {
		return nil, &ApplicationError{
			HttpCode:     http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		}
	}
	res, err := result.LastInsertId()
	if err != nil {
		return nil, &ApplicationError{
			HttpCode:     http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		}
	}
	record := make(map[string]interface{})
	record[h.GetFieldNamePrimaryKey(tableName)] = res
	return record, nil
}

func (h *Handler) CreateQueryUpdateById(tableName string, fData map[string]FrontData, id string) (string, Row, error) {
	var q strings.Builder
	var qEnd strings.Builder
	q.WriteString(fmt.Sprintf("UPDATE %s SET ", tableName))
	fieldPrimary := h.GetFieldNamePrimaryKey(tableName)
	zapytay := false
	rowValue := make(Row, 0)
	insert := false
	var idP interface{}
	for index, row := range h.Table[tableName] {
		insert = false
		field := fmt.Sprintf("%s", row.Field)
		key := fmt.Sprintf("%s", row.Key)
		tField := fmt.Sprintf("%s", row.Type)

		if field == fieldPrimary {
			if strings.Contains(tField, "int") {
				idInt, err := strconv.Atoi(id)
				if err != nil {
					return "", nil, &ApplicationError{
						HttpCode:     http.StatusBadRequest,
						ErrorMessage: fmt.Sprintf("field %s have invalid type", field),
					}
				}
				idP = idInt
			} else if strings.Contains(tField, "float") || strings.Contains(tField, "double") {
				idFloat, err := strconv.ParseFloat(id, 64)
				if err != nil {
					return "", nil, &ApplicationError{
						HttpCode:     http.StatusBadRequest,
						ErrorMessage: fmt.Sprintf("field %s have invalid type", field),
					}
				}
				idP = idFloat
			}
		}

		val, ok := fData[field]
		if val.Key == "" {
			continue
		}

		if key == "PRI" && ok {
			return "", nil, &ApplicationError{
				HttpCode:     http.StatusBadRequest,
				ErrorMessage: fmt.Sprintf("field %s have invalid type", field),
			}
		}

		if zapytay {
			q.WriteString(", ")
			qEnd.WriteString(", ")
		}

		if strings.Contains(tField, "int") {
			valInt, err := strconv.Atoi(fmt.Sprintf("%s", val.Value))
			if err != nil {
				return "", nil, &ApplicationError{
					HttpCode:     http.StatusBadRequest,
					ErrorMessage: fmt.Sprintf("field %s have invalid type", field),
				}
			}
			rowValue = append(rowValue, valInt)
			insert = true
		} else if strings.Contains(tField, "float") || strings.Contains(tField, "double") {
			valFloat, ok := val.Value.(float64)
			if !ok {
				return "", nil, &ApplicationError{
					HttpCode:     http.StatusBadRequest,
					ErrorMessage: fmt.Sprintf("field %s have invalid type", field),
				}
			}
			rowValue = append(rowValue, valFloat)
			insert = true
		}

		if _, ok1 := val.Value.(string); (strings.Contains(tField, "text") || strings.Contains(tField, "varchar")) && !ok1 && val.Value != nil {
			return "", nil, &ApplicationError{
				HttpCode:     http.StatusBadRequest,
				ErrorMessage: fmt.Sprintf("field %s have invalid type", field),
			}
		}

		if strings.Contains(fmt.Sprintf("%s", row.Null), "NO") && val.Value == nil {
			return "", nil, &ApplicationError{
				HttpCode:     http.StatusBadRequest,
				ErrorMessage: fmt.Sprintf("field %s have invalid type", field),
			}
		}

		if index == len(h.Table[tableName])-1 {
			q.WriteString(fmt.Sprintf("%s = ? ", row.Field))
			if !insert {
				rowValue = append(rowValue, val.Value)
			}
			continue
		}

		q.WriteString(fmt.Sprintf("%s = ?", row.Field))

		if !insert {
			rowValue = append(rowValue, val.Value)
		}
		zapytay = true
	}
	if idP == nil {
		rowValue = append(rowValue, id)
	} else {
		rowValue = append(rowValue, idP)
	}

	q.WriteString(fmt.Sprintf(" WHERE %s = ?", fieldPrimary))
	return q.String(), rowValue, nil
}

func (h *Handler) StorageUpdateObject(tableName string, fData map[string]FrontData, id string) (map[string]interface{}, error) {
	q, rowVal, err := h.CreateQueryUpdateById(tableName, fData, id)
	if err != nil {
		return nil, err
	}
	fmt.Printf("q: %s, rowVal: %v, err: %s\n", q, rowVal, err)
	result, err := h.DB.Exec(q, rowVal...)
	if err != nil {
		return nil, &ApplicationError{
			HttpCode:     http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		}
	}
	res, err := result.RowsAffected()
	if err != nil {
		return nil, &ApplicationError{
			HttpCode:     http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		}
	}
	resp := make(map[string]interface{})
	resp["updated"] = res
	fmt.Printf("%v\n", resp)
	return resp, nil
}

func (h *Handler) StorageGetById(tableName string, reqId string) (map[string]interface{}, error) {
	rows := h.Table[tableName]
	q := h.CreateQueryGetById(tableName)
	result, err := h.DB.Query(q, reqId)

	if err != nil {
		return nil, &ApplicationError{
			HttpCode:     http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		}
	}
	record := make(map[string]interface{})
	for result.Next() {
		args := make(Row, len(rows))
		argsMap := make(map[string]interface{})

		for i := 0; i < len(args); i++ {
			args[i] = new(interface{})
		}

		err := result.Scan(args...)
		if err != nil {
			return nil, &ApplicationError{
				HttpCode:     http.StatusInternalServerError,
				ErrorMessage: err.Error(),
			}
		}
		for i := 0; i < len(args); i++ {
			ll := *args[i].(*interface{})
			t := fmt.Sprintf("%s", rows[i].Type)
			if (strings.Contains(t, "text") || strings.Contains(t, "varchar")) && ll != nil {
				args[i] = string(ll.([]uint8))
			}
			key := fmt.Sprintf("%s", rows[i].Field)
			argsMap[key] = args[i]
		}
		record["record"] = argsMap
	}
	result.Close()
	if len(record) == 0 {
		return nil, &ApplicationError{
			HttpCode:     http.StatusNotFound,
			ErrorMessage: "record not found",
		}
	}

	return record, nil
}

func (h *Handler) StorageGetRecords(tableName string, param *Param) (map[string]interface{}, error) {
	rows := h.Table[tableName]
	q := h.CreateQueryGetRecords(tableName, param)
	result, err := h.DB.Query(q)
	if err != nil {
		return nil, &ApplicationError{
			HttpCode:     http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		}
	}

	recordsMap := make(map[string]interface{})
	records := make([]map[string]interface{}, 0)
	for result.Next() {
		args := make(Row, len(rows))
		argsMap := make(map[string]interface{})
		for i := 0; i < len(args); i++ {
			args[i] = new(interface{})
		}

		err := result.Scan(args...)
		if err != nil {
			return nil, &ApplicationError{
				HttpCode:     http.StatusInternalServerError,
				ErrorMessage: err.Error(),
			}
		}
		for i := 0; i < len(args); i++ {
			key := fmt.Sprintf("%s", rows[i].Field)
			ll := *args[i].(*interface{})
			t := fmt.Sprintf("%s", rows[i].Type)
			if (strings.Contains(t, "text") || strings.Contains(t, "varchar")) && ll != nil {
				args[i] = string(ll.([]uint8))
			}
			argsMap[key] = args[i]
		}
		records = append(records, argsMap)
	}
	result.Close()
	if len(records) == 0 {
		return nil, &ApplicationError{
			HttpCode:     http.StatusNotFound,
			ErrorMessage: "Object not found",
		}
	}
	recordsMap["records"] = records
	return recordsMap, nil
}

func (h *Handler) GetById(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")

	if resp, err := h.StorageGetById(path[1], path[2]); err != nil {
		if errApp, ok := err.(*ApplicationError); ok {
			w.WriteHeader(errApp.HttpCode)
			resp := &Response{
				"error": errApp.ErrorMessage,
			}
			errAppJson, _ := json.Marshal(resp)
			w.Write(errAppJson)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		errJson, _ := json.Marshal(err)
		w.Write(errJson)
	} else {
		w.WriteHeader(http.StatusOK)
		resp := &Response{
			"response": resp,
		}
		respJson, _ := json.Marshal(resp)
		w.Write(respJson)
	}
}

func (h *Handler) GetRecordsHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", err)
		return
	}
	param := &Param{
		Limit:  10,
		Offset: 0,
	}
	if limitStr := r.Form.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil && limit > 0 {
			param.Limit = limit
		}
	}
	if offsetStr := r.Form.Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err == nil && offset > -1 {
			param.Offset = offset
		}
	}

	if resp, err := h.StorageGetRecords(path[1], param); err != nil {
		if errApp, ok := err.(*ApplicationError); ok {
			w.WriteHeader(errApp.HttpCode)
			errAppJson, _ := json.Marshal(errApp)
			w.Write(errAppJson)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		errJson, _ := json.Marshal(err)
		w.Write(errJson)
	} else {
		response := &Response{
			"response": resp,
		}
		w.WriteHeader(http.StatusOK)
		responseJson, _ := json.Marshal(response)
		w.Write(responseJson)
	}
}

func (h *Handler) DeleteById(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")

	if res, err := h.StorageDeleteById(path[1], path[2]); err != nil {
		if errApp, ok := err.(*ApplicationError); ok {
			w.WriteHeader(errApp.HttpCode)
			resp := &Response{
				"error": errApp.ErrorMessage,
			}
			respJson, _ := json.Marshal(resp)
			w.Write(respJson)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		resp := &Response{
			"error": err.Error(),
		}
		respJson, _ := json.Marshal(resp)
		w.Write(respJson)
	} else {
		w.WriteHeader(http.StatusOK)
		resp := &Response{
			"response": res,
		}
		respJson, _ := json.Marshal(resp)
		w.Write(respJson)
	}
}

func (h *Handler) CreateObjectHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	frontDataMap := make(map[string]FrontData)
	tableName := path[1]
	infoTable := h.Table[tableName]
	decoder := json.NewDecoder(r.Body)
	kek := make(map[string]interface{})
	decoder.Decode(&kek)

	for _, info := range infoTable {
		field := strings.ToLower(fmt.Sprintf("%s", info.Field))
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "%s", err)
			return
		}
		if _, ok := kek[field]; ok {
			d := &FrontData{
				Key:   fmt.Sprintf("%s", info.Field),
				Value: kek[field],
			}
			frontDataMap[d.Key] = *d
		}
	}

	if resp, err := h.StorageInsertObject(tableName, frontDataMap); err != nil {
		if errApp, ok := err.(*ApplicationError); ok {
			w.WriteHeader(errApp.HttpCode)
			resp := &Response{
				"error": errApp.ErrorMessage,
			}
			respJson, _ := json.Marshal(resp)
			w.Write(respJson)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		respJson, _ := json.Marshal(resp)
		w.Write(respJson)
		w.Write([]byte(err.Error()))

	} else {
		w.WriteHeader(http.StatusOK)
		resp := &Response{
			"response": resp,
		}
		respJson, _ := json.Marshal(resp)
		w.Write(respJson)
	}
}

func (h *Handler) UpdateObjectHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	frontDataMap := make(map[string]FrontData)
	tableName := path[1]
	id := path[2]
	infoTable := h.Table[tableName]

	decoder := json.NewDecoder(r.Body)
	kek := make(map[string]interface{})
	decoder.Decode(&kek)

	for _, info := range infoTable {
		field := strings.ToLower(fmt.Sprintf("%s", info.Field))
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "%s", err)
			return
		}
		if _, ok := kek[field]; ok {
			d := &FrontData{
				Key:   fmt.Sprintf("%s", info.Field),
				Value: kek[field],
			}
			frontDataMap[d.Key] = *d
		}
	}

	if res, err := h.StorageUpdateObject(tableName, frontDataMap, id); err != nil {
		if errApp, ok := err.(*ApplicationError); ok {
			w.WriteHeader(errApp.HttpCode)
			resp := &Response{
				"error": errApp.ErrorMessage,
			}
			respJson, _ := json.Marshal(resp)
			w.Write(respJson)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)

		resp := &Response{
			"error": err.Error(),
		}
		respJson, _ := json.Marshal(resp)
		w.Write(respJson)

	} else {
		w.WriteHeader(http.StatusOK)
		resp := &Response{
			"response": res,
		}
		respJson, _ := json.Marshal(resp)
		w.Write(respJson)
	}
}

func (h *Handler) GetFieldNamePrimaryKey(tableName string) string {
	for _, row := range h.Table[tableName] {
		key := fmt.Sprintf("%s", row.Key)
		if key == "PRI" {
			return fmt.Sprintf("%s", row.Field)
		}
	}
	return ""
}

// func (h *Handler) AddFilters(query *strings.Builder, limit int, offset int) *strings.Builder {
// 	if limit != -1 {
// 		query.WriteString(fmt.Sprintf(" LIMIT ?", limit))
// 	}
// 	if offset != -1 {
// 		query.WriteString(fmt.Sprintf(" OFFSET ?", offset))
// 	}
// 	return ""
// }

func (h *Handler) InfoTables() error {
	h.Table = make(map[string][]InfoTable)
	result, err := h.DB.Query("SHOW TABLES")

	if err != nil {
		return ApplicationError{
			HttpCode:     http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		}
	}
	for result.Next() {
		var tableName string
		if err := result.Scan(&tableName); err != nil {
			return ApplicationError{
				HttpCode:     http.StatusInternalServerError,
				ErrorMessage: err.Error(),
			}
		}

		h.Table[tableName] = make([]InfoTable, 0)
	}
	result.Close()

	for tableName := range h.Table {

		q := fmt.Sprintf("SHOW FULL COLUMNS FROM `%s`", tableName)
		res, err := h.DB.Query(q)

		if err != nil {
			return ApplicationError{
				HttpCode:     http.StatusInternalServerError,
				ErrorMessage: err.Error(),
			}
		}
		infoTable := new(InfoTable)
		// infoTableArr := make([]InfoTable, 0)
		for res.Next() {
			if err := res.Scan(&infoTable.Field, &infoTable.Type,
				&infoTable.Collation, &infoTable.Null,
				&infoTable.Key, &infoTable.Default,
				&infoTable.Extra, &infoTable.Privileges, &infoTable.Comment); err != nil {
				return ApplicationError{
					HttpCode:     http.StatusInternalServerError,
					ErrorMessage: err.Error(),
				}
			}
			h.Table[tableName] = append(h.Table[tableName], *infoTable)

		}
		res.Close()
	}
	return nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "application/json")
	fmt.Printf("path: %s\n", r.URL.Path)

	if r.URL.Path == "/" {
		resp := make(Response)
		ans := make(Response)
		resp["response"] = &ans
		tables := make([]interface{}, 0)

		for table := range h.Table {
			tables = append(tables, table)
		}
		ans["tables"] = tables
		respJson, _ := json.Marshal(resp)
		fmt.Fprintf(w, "%s", respJson)
		w.WriteHeader(http.StatusOK)
		return
	}

	for table := range h.Table {
		path := strings.Split(r.URL.Path, "/")
		if table == path[1] && len(path) == 3 {
			switch r.Method {
			case http.MethodGet:
				h.GetById(w, r)
			case http.MethodDelete:
				h.DeleteById(w, r)
			case http.MethodPost:
				h.UpdateObjectHandler(w, r)
			case http.MethodPut:
				h.CreateObjectHandler(w, r)
			}
			return
		} else if table == path[1] && len(path) == 2 {
			if r.Method == http.MethodPut {
				h.CreateObjectHandler(w, r)
				return
			}
			if r.Method == http.MethodGet {
				h.GetRecordsHandler(w, r)
				return
			}
		}
	}

	w.WriteHeader(http.StatusNotFound)
	resp := make(Response)
	resp["error"] = "unknown table"
	errJson, _ := json.Marshal(resp)
	fmt.Fprintf(w, "%s", errJson)
}

func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	h := &Handler{
		DB: db,
	}
	if err := h.InfoTables(); err != nil {
		return nil, fmt.Errorf("Error get info about tables database.\n")
	}
	// fmt.Printf("%s\n", h.Table)
	// fmt.Printf("%v\n", h.DB.Stats().OpenConnections)
	return h, nil
}

type ApplicationError struct {
	HttpCode     int
	ErrorMessage string
}

func (e ApplicationError) Error() string {
	return fmt.Sprintf("error: %s", e.ErrorMessage)
}
