package auth

import models "github.com/tubes-cc/logistics/domain"

// UserRepository adalah interface (kontrak)
type UserRepository interface {
	GetByEmail(email string) (*models.User, error)
	Create(user *models.User) error
}