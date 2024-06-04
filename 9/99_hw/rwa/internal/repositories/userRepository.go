package repositories

import (
	"fmt"
	"rwa/internal/models"
	"sync"
)

type UserRepository struct {
	mu       sync.RWMutex
	users    map[uint]*models.User
	sequence uint
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		mu:       sync.RWMutex{},
		users:    make(map[uint]*models.User),
		sequence: 0,
	}
}

func (ur *UserRepository) Add(user *models.User) (*models.User, error) {
	ur.mu.Lock()
	defer ur.mu.Unlock()
	ur.sequence++
	user.ID = ur.sequence
	ur.users[ur.sequence] = user
	return user, nil
}

func (ur *UserRepository) Get(userID uint) (*models.User, error) {
	ur.mu.RLock()
	defer ur.mu.RUnlock()
	user, ok := ur.users[userID]
	if !ok {
		return nil, fmt.Errorf("User not found")
	}
	return user, nil
}

func (ur *UserRepository) Delete(userID uint) bool {
	ur.mu.Lock()
	defer ur.mu.Unlock()
	before := len(ur.users)
	delete(ur.users, userID)
	after := len(ur.users)
	return before > after
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
