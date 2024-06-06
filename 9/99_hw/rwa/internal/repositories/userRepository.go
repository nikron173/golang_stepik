package repositories

import (
	"fmt"
	"log"
	"rwa/internal/common"
	"rwa/internal/models"
	"sync"
	"time"
)

var (
	userRepository *UserRepository
)

type UserRepository struct {
	mu    sync.RWMutex
	users map[string]*models.User
}

func NewUserRepository() *UserRepository {
	if userRepository == nil {
		userRepository = &UserRepository{
			mu:    sync.RWMutex{},
			users: make(map[string]*models.User),
		}
	}
	return userRepository
}

func (ur *UserRepository) Add(user *models.User) (*models.User, error) {
	ur.mu.Lock()
	defer ur.mu.Unlock()
	user.ID = common.RandStringRunes(15)
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt
	ur.users[user.ID] = user
	return user, nil
}

func (ur *UserRepository) Get(userID string) (*models.User, error) {
	ur.mu.RLock()
	defer ur.mu.RUnlock()
	user, ok := ur.users[userID]
	if !ok {
		return nil, fmt.Errorf("User not found")
	}
	return user, nil
}

func (ur *UserRepository) Delete(userID string) bool {
	ur.mu.Lock()
	defer ur.mu.Unlock()
	before := len(ur.users)
	delete(ur.users, userID)
	after := len(ur.users)
	return before > after
}

func (ur *UserRepository) Update(userID string, user *models.User) (*models.User, error) {
	ur.mu.Lock()
	defer ur.mu.Unlock()
	userDB, ok := ur.users[userID]
	if !ok {
		return nil, fmt.Errorf("User not found db")
	}
	update := false
	if user.BIO != "" {
		userDB.BIO = user.BIO
		update = true
	}
	if user.Email != "" {
		userDB.Email = user.Email
		update = true
	}
	if user.Password != "" {
		userDB.Password = user.Password
		update = true
	}
	if update {
		userDB.UpdatedAt = time.Now()
	}
	log.Printf("UserRepository: Update: userDB: %#v\n", userDB)
	return userDB, nil
}

func (ur *UserRepository) GetByPasswordAndEmail(password string, email string) (*models.User, bool) {
	ur.mu.RLock()
	defer ur.mu.RUnlock()
	for _, user := range ur.users {
		if user.Email == email && user.Password == password {
			return user, true
		}
	}
	return nil, false
}
