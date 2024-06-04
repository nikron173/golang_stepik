package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"rwa/internal/handlers/middleware"
	"rwa/internal/models"
	"rwa/internal/repositories"
	"rwa/internal/service"
	"time"
)

type Storage interface {
	Add(*models.User) (*models.User, error)
	Get(uint) (*models.User, error)
	GetByPasswordAndEmail(password string, email string) (*models.User, bool)
	Delete(uint) bool
}

type UserHandler struct {
	storage Storage
	sm      *service.SessionService
}

func NewUserHandler(sm *service.SessionService) *UserHandler {
	return &UserHandler{
		storage: repositories.NewUserRepository(),
		sm:      sm,
	}
}

func (uh *UserHandler) Add(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user := &models.User{}
	json.Unmarshal(body, user)
	uh.storage.Add(user)
	userJson, err := json.Marshal(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(userJson)
}

func (uh *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	session, err := middleware.GetSessionFromContext(r.Context())
	log.Printf("UserHandler: Get: session: %#v\n", session)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	user, err := uh.storage.Get(session.UserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userJson, err := json.Marshal(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(userJson)
}

func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user := &models.User{}
	err = json.Unmarshal(body, user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userDB, ok := uh.storage.GetByPasswordAndEmail(user.Password, user.Email)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	session, err := uh.sm.Create(userDB.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	cokkie := http.Cookie{
		Name:    "Authorization",
		Value:   session.ID,
		Expires: time.Now().Add(time.Hour * 24),
	}

	http.SetCookie(w, &cokkie)

	userJson, err := json.Marshal(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(userJson)
}

func (uh *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("Authorization")
	if err == http.ErrNoCookie {
		w.WriteHeader(http.StatusOK)
		return
	}
	uh.sm.Delete(session.Value)
	session.Expires = time.Now().AddDate(0, 0, -1)
	http.SetCookie(w, session)
	w.WriteHeader(http.StatusOK)
}

func (uh *UserHandler) Modify(w http.ResponseWriter, r *http.Request) {

}

func (uh *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/users":
		{
			switch r.Method {
			case http.MethodPost:
				uh.Add(w, r)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		}
	case "/api/users/login":
		{
			switch r.Method {
			case http.MethodPost:
				uh.Login(w, r)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		}
	case "/api/user":
		{
			switch r.Method {
			case http.MethodGet:
				uh.Get(w, r)
			case http.MethodPut:
				uh.Modify(w, r)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		}
	case "/api/user/logout":
		{
			switch r.Method {
			case http.MethodPost:
				uh.Logout(w, r)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
