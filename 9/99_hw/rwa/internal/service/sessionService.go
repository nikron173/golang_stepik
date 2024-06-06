package service

import (
	"fmt"
	"log"
	"net/http"
	"rwa/internal/models"
	"rwa/internal/repositories"
	"strings"
)

type Storage interface {
	Check(string) (*models.Session, bool)
	Delete(string) bool
	Create(string) *models.Session
	DeleteAll(string)
}

type SessionService struct {
	sessionStorage Storage
}

func NewSessionService() *SessionService {
	return &SessionService{
		sessionStorage: repositories.NewSessionRepository(),
	}
}

func (ss *SessionService) Check(r *http.Request) (*models.Session, error) {
	// token, err := r.Cookie("Authorization")
	token := r.Header.Get("Authorization")

	// if err == http.ErrNoCookie {
	// 	return nil, fmt.Errorf("No auth")
	// }
	log.Printf("SessionService: Check: token: %#v\n", token)
	if token == "" || len(strings.Split(token, " ")) < 2 {
		return nil, fmt.Errorf("No auth")
	}

	// session, ok := ss.sessionStorage.Check(token.Value)
	session, ok := ss.sessionStorage.Check(strings.Split(token, " ")[1])
	log.Println(session)
	if ok {
		return session, nil
	}
	return nil, fmt.Errorf("Session not found")
}

func (ss *SessionService) Create(userID string) (*models.Session, error) {
	return ss.sessionStorage.Create(userID), nil
}

func (ss *SessionService) Delete(token string) bool {
	return ss.sessionStorage.Delete(token)
}

func (ss *SessionService) DeleteAll(userID string) {
	ss.sessionStorage.DeleteAll(userID)
}
