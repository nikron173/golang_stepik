package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"rwa/internal/models"
)

type ctxKey string

var (
	publicUrlPath = map[string]struct{}{
		"/api/articles":    struct{}{},
		"/api/users":       struct{}{},
		"/api/users/login": struct{}{},
	}
	sessionID ctxKey = "Authorization"
)

type SessionManager interface {
	Check(*http.Request) (*models.Session, error)
}

func GetSessionFromContext(cxt context.Context) (*models.Session, error) {
	session, ok := cxt.Value(sessionID).(*models.Session)
	if !ok {
		return nil, fmt.Errorf("No auth")
	}
	return session, nil
}

func AuthMiddleware(sm SessionManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		if _, ok := publicUrlPath[r.URL.Path]; ok {
			next.ServeHTTP(w, r)
			return
		}

		session, err := sm.Check(r)
		log.Printf("Middleware: session: %#v, err: %s\n", session, err)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), sessionID, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
