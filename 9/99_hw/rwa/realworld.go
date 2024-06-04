package main

import (
	"net/http"
	"rwa/internal/handlers"
	"rwa/internal/handlers/middleware"
	"rwa/internal/service"
)

// сюда писать код

func GetApp() http.Handler {

	sm := service.NewSessionService()
	userHandler := handlers.NewUserHandler(sm)

	mux := http.NewServeMux()
	mux.Handle("/api/users", userHandler)
	mux.Handle("/api/users/login", userHandler)
	mux.Handle("/api/user", userHandler)

	auth := middleware.AuthMiddleware(sm, mux)

	return auth
}
