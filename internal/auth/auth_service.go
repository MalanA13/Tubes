package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	models "github.com/tubes-cc/logistics/domain"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo UserRepository
}

func NewAuthService(r UserRepository) *AuthService {
	return &AuthService{repo: r}
}

// Register creates a new user account with email and password.
func (s *AuthService) Register(fullName, email, password string) (*models.User, error) {
	if email == "" || password == "" {
		return nil, errors.New("email and password are required")
	}

	if len(password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	// Check if user already exists
	existing, err := s.repo.GetByEmail(email)
	if err == nil && existing != nil {
		return nil, errors.New("user already exists")
	}

	// Hash password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		FullName:  fullName,
		Email:     email,
		Password:  string(hashedPassword),
		Role:      "customer", // Default role
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login authenticates user and returns JWT token.
func (s *AuthService) Login(email, password string) (string, error) {
	if email == "" || password == "" {
		return "", errors.New("email and password are required")
	}

	// Fetch user by email
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return "", errors.New("user not found")
	}

	if user == nil {
		return "", errors.New("user not found")
	}

	// Verify password using bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("incorrect password")
	}

	// Generate JWT token
	token, err := s.generateToken(user)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

// ValidateToken verifies a JWT token and returns user claims.
func (s *AuthService) ValidateToken(tokenString string) (*models.AuthClaims, error) {
	if tokenString == "" {
		return nil, errors.New("token is required")
	}

	token, err := jwt.ParseWithClaims(tokenString, &models.AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(s.getJWTSecret()), nil
	})

	if err != nil {
		return nil, fmt.Errorf("token parsing failed: %w", err)
	}

	claims, ok := token.Claims.(*models.AuthClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("token has expired")
	}

	return claims, nil
}

// generateToken creates a new JWT token for a user.
func (s *AuthService) generateToken(user *models.User) (string, error) {
	now := time.Now()
	expiryTime := now.Add(24 * time.Hour) // Token valid for 24 hours

	claims := &models.AuthClaims{
		UserID: fmt.Sprintf("%d", user.ID),
		Role:   models.UserRole(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiryTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "auth-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.getJWTSecret()))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// getJWTSecret retrieves JWT secret from environment or uses default.
func (s *AuthService) getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-secret-key-change-in-production" // CHANGE THIS IN PRODUCTION
	}
	return secret
}