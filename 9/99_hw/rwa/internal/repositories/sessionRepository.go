package repositories

import (
	"log"
	"rwa/internal/common"
	"rwa/internal/models"
	"sync"
)

type SessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]*models.Session
}

func NewSessionRepository() *SessionRepository {
	return &SessionRepository{
		mu:       sync.RWMutex{},
		sessions: make(map[string]*models.Session),
	}
}

func (sr *SessionRepository) Create(userID string) *models.Session {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sessionID := common.RandStringRunes(16)
	session := &models.Session{
		ID:     sessionID,
		UserID: userID,
	}
	sr.sessions[sessionID] = session
	log.Printf("Create session: %#v\n", session)
	return session
}

func (sr *SessionRepository) Check(token string) (*models.Session, bool) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	log.Printf("SessionRepository: Map sessions: %#v\n", sr.sessions)
	session, ok := sr.sessions[token]
	return session, ok
}

func (sr *SessionRepository) Delete(token string) bool {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	before := len(sr.sessions)
	delete(sr.sessions, token)
	after := len(sr.sessions)
	return before > after
}

func (sr *SessionRepository) DeleteAll(userID string) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sliceSessionDelete := make([]string, 0)
	for _, session := range sr.sessions {
		if session.UserID == userID {
			sliceSessionDelete = append(sliceSessionDelete, session.ID)
		}
	}
	for _, sessionID := range sliceSessionDelete {
		delete(sr.sessions, sessionID)
	}
}
