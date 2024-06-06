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
	Get(string) (*models.User, error)
	GetByPasswordAndEmail(password string, email string) (*models.User, bool)
	Delete(string) bool
	Update(string, *models.User) (*models.User, error)
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

	mapUser := make(map[string]*models.User)
	json.Unmarshal(body, &mapUser)
	user, ok := mapUser["user"]

	if !ok {
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	userDB, err := uh.storage.Add(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("UserHandler: Add: %#v\n", userDB)

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

	resp := make(map[string]*models.User)
	resp["User"] = &models.User{
		Username: userDB.Username,
		Email:    userDB.Email,
		Token:    session.ID,
	}
	respJson, err := json.Marshal(resp)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// log.Printf("UserHandler: Add: %#v\n", string(respJson))
	w.WriteHeader(http.StatusCreated)
	w.Write(respJson)
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

	resp := make(map[string]*models.User)
	resp["User"] = &models.User{
		Username: user.Username,
		Email:    user.Email,
		BIO:      user.BIO,
	}

	respJson, err := json.Marshal(resp)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respJson)
}

func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	mapUser := make(map[string]*models.User)
	json.Unmarshal(body, &mapUser)
	user, ok := mapUser["user"]

	if !ok {
		w.WriteHeader(http.StatusBadGateway)
		return
	}

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

	resp := make(map[string]*models.User)
	resp["User"] = &models.User{
		Username: userDB.Username,
		Email:    userDB.Email,
		Token:    session.ID,
	}

	respJson, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(respJson)
}

func (uh *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session, err := middleware.GetSessionFromContext(r.Context())
	log.Printf("UserHandler: Get: session: %#v\n", session)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	uh.sm.Delete(session.ID)
	w.WriteHeader(http.StatusOK)
}

func (uh *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	session, err := middleware.GetSessionFromContext(r.Context())
	log.Printf("UserHandler: Get: session: %#v\n", session)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	mapUser := make(map[string]*models.User)
	json.Unmarshal(body, &mapUser)
	user, ok := mapUser["user"]
	if !ok {
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	userDB, err := uh.storage.Update(session.UserID, user)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	uh.sm.Delete(session.ID)
	sessionNew, err := uh.sm.Create(userDB.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := make(map[string]*models.User)
	resp["User"] = &models.User{
		Username: userDB.Username,
		Email:    userDB.Email,
		BIO:      userDB.BIO,
		Token:    sessionNew.ID,
	}

	respJson, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("UserHandler: Update: respJson: %#v\n", string(respJson))
	w.WriteHeader(http.StatusOK)
	w.Write(respJson)
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
				uh.Update(w, r)
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
