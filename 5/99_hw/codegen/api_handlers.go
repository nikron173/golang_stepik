package main

import (
		"encoding/json"
		"fmt"
		"net/http"
		"strconv"
)


func (h *OtherApi) handlerOtherApiCreate(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotAcceptable)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}", "error", "bad method")
		return
	}


	if r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}", "error", "unauthorized")
		return
	}
				
	in := new(OtherCreateParams)

	usernameStr := r.FormValue("username")
	
	if usernameStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}",


			"error", "username must me not empty")
		return
	}

	username := usernameStr
	if len(username) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}",

			"error", "username len must be >= 3")
		return
	}
	in.Username = username

	nameStr := r.FormValue("account_name")


	name := nameStr
	in.Name = name

	classStr := r.FormValue("class")
	

	class := classStr

	if class == "" {
		class = "warrior"
	}
	if !("warrior" == class || "sorcerer" == class || "rouge" == class) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}",

			"error", "class must be one of [warrior, sorcerer, rouge]")
		return
	}
	in.Class = class

	levelStr := r.FormValue("level")
	
	level, err := strconv.Atoi(levelStr)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}",

			"error", "level must be int")
		return
	}
	if level < 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}",

			"error", "level must be >= 1")
		return
	}
	if level > 50 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}",

			"error", "level must be <= 50")
		return
	}
	in.Level = level
	res, err := h.Create(r.Context(), *in)
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

func (h *MyApi) handlerMyApiProfile(w http.ResponseWriter, r *http.Request) {
	in := new(ProfileParams)

	loginStr := r.FormValue("login")
	
	if loginStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}",


			"error", "login must me not empty")
		return
	}

	login := loginStr
	in.Login = login
	res, err := h.Profile(r.Context(), *in)
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

func (h *MyApi) handlerMyApiCreate(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotAcceptable)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}", "error", "bad method")
		return
	}


	if r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}", "error", "unauthorized")
		return
	}
				
	in := new(CreateParams)

	loginStr := r.FormValue("login")
	
	if loginStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}",


			"error", "login must me not empty")
		return
	}

	login := loginStr
	if len(login) < 10 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}",

			"error", "login len must be >= 10")
		return
	}
	in.Login = login

	nameStr := r.FormValue("full_name")


	name := nameStr
	in.Name = name

	statusStr := r.FormValue("status")
	

	status := statusStr

	if status == "" {
		status = "user"
	}
	if !("user" == status || "moderator" == status || "admin" == status) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}",

			"error", "status must be one of [user, moderator, admin]")
		return
	}
	in.Status = status

	ageStr := r.FormValue("age")
	
	age, err := strconv.Atoi(ageStr)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}",

			"error", "age must be int")
		return
	}
	if age < 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}",

			"error", "age must be >= 0")
		return
	}
	if age > 128 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}",

			"error", "age must be <= 128")
		return
	}
	in.Age = age
	res, err := h.Create(r.Context(), *in)
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

func (h *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "application/json")
	switch r.URL.Path {
	case "/user/profile":
		h.handlerMyApiProfile(w, r)
	case "/user/create":
		h.handlerMyApiCreate(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}",
			"error", "unknown method")
	}
}
		
func (h *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "application/json")
	switch r.URL.Path {
	case "/user/create":
		h.handlerOtherApiCreate(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "{\"%s\": \"%s\"}",
			"error", "unknown method")
	}
}
		
