package handlers

import (
	"net/http"
	"rwa/internal/models"
	"rwa/internal/repositories"
)

type Storage interface {
	Add(*models.User) (*models.User, error)
}

type UserHandler struct {
	storage Storage
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		storage: repositories.NewUserRepository(),
	}
}

func (uh *UserHandler) Add(w http.ResponseWriter, r *http.Request) {

}

func (uh *UserHandler) Get(w http.ResponseWriter, r *http.Request) {

}

func (uh *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
