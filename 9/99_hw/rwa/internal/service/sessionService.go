package service

import (
	"fmt"
	"log"
	"net/http"
	"rwa/internal/models"
	"rwa/internal/repositories"
)

type Storage interface {
	Check(string) (*models.Session, bool)
	Delete(string) bool
	Create(uint) *models.Session
	DeleteAll(uint)
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
	token, err := r.Cookie("Authorization")

	if err == http.ErrNoCookie {
		return nil, fmt.Errorf("No auth")
	}

	session, ok := ss.sessionStorage.Check(token.Value)
	log.Println(session)
	if ok {
		return session, nil
	}
	return nil, fmt.Errorf("Session not found")
}

func (ss *SessionService) Create(userID uint) (*models.Session, error) {
	return ss.sessionStorage.Create(userID), nil
}

func (ss *SessionService) Delete(token string) bool {
	return ss.sessionStorage.Delete(token)
}

func (ss *SessionService) DeleteAll(userID uint) {
	ss.sessionStorage.DeleteAll(userID)
}
