package main

import (
	"net/http"
	"rwa/internal/handlers"
)

// сюда писать код

func GetApp() http.Handler {

	userHandler := handlers.NewUserHandler()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/users", userHandler.Add)
	mux.Handle("/api/users/login", userHandler)
	mux.HandleFunc("/api/user", userHandler.Get)

	return nil
}
