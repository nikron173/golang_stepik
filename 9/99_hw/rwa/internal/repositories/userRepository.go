package repositories

import "rwa/internal/models"

type UserRepository struct {
}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (ur *UserRepository) Add(user *models.User) (*models.User, error) {

	return nil, nil
}
