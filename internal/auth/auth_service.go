package auth

import (
	"errors"
)

type AuthService struct {
	repo UserRepository
}

func NewAuthService(r UserRepository) *AuthService {
	return &AuthService{repo: r}
}

func (s *AuthService) Login(email, password string) (string, error) {
	// 1. Cari user berdasarkan email lewat repository
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return "", errors.New("user tidak ditemukan")
	}

	// 2. Cek password (masih sederhana dulu)
	if user.Password != password {
		return "", errors.New("password salah")
	}

	// 3. Return token (ceritanya sukses)
	return "TOKEN_JWT_DUMMY", nil
}