package auth

import (
	"strings"
	"testing"
	"time"

	models "github.com/tubes-cc/logistics/domain"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockUserRepository(ctrl)
	authService := NewAuthService(mockRepo)

	// Mock GetByEmail to return nil (user doesn't exist)
	mockRepo.EXPECT().GetByEmail("newuser@example.com").Return(nil, nil)

	// Mock Create to return nil (success)
	mockRepo.EXPECT().Create(gomock.Any()).DoAndReturn(func(user *models.User) error {
		assert.Equal(t, "newuser@example.com", user.Email)
		assert.Equal(t, "John Doe", user.FullName)
		assert.Equal(t, "customer", user.Role)
		// Verify password is hashed
		assert.NotEqual(t, "password123", user.Password)
		return nil
	})

	user, err := authService.Register("John Doe", "newuser@example.com", "password123")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "newuser@example.com", user.Email)
	assert.Equal(t, "John Doe", user.FullName)
}

func TestRegister_UserAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockUserRepository(ctrl)
	authService := NewAuthService(mockRepo)

	existingUser := &models.User{Email: "existing@example.com"}
	mockRepo.EXPECT().GetByEmail("existing@example.com").Return(existingUser, nil)

	user, err := authService.Register("John", "existing@example.com", "password123")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "user already exists", err.Error())
}

func TestRegister_WeakPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockUserRepository(ctrl)
	authService := NewAuthService(mockRepo)

	user, err := authService.Register("John", "test@example.com", "123")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "password must be at least 6 characters", err.Error())
}

func TestLogin_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockUserRepository(ctrl)
	authService := NewAuthService(mockRepo)

	// Hash a password for the test user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	testUser := &models.User{
		ID:        1,
		Email:     "test@example.com",
		Password:  string(hashedPassword),
		FullName:  "Test User",
		Role:      "customer",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.EXPECT().GetByEmail("test@example.com").Return(testUser, nil)

	token, err := authService.Login("test@example.com", "password123")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	// Verify token is a JWT (should have 3 parts separated by dots)
	parts := strings.Split(token, ".")
	assert.Equal(t, 3, len(parts))
}

func TestLogin_InvalidPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockUserRepository(ctrl)
	authService := NewAuthService(mockRepo)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testUser := &models.User{
		Email:    "test@example.com",
		Password: string(hashedPassword),
		Role:     "customer",
	}

	mockRepo.EXPECT().GetByEmail("test@example.com").Return(testUser, nil)

	token, err := authService.Login("test@example.com", "wrongpassword")
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Equal(t, "incorrect password", err.Error())
}

func TestLogin_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockUserRepository(ctrl)
	authService := NewAuthService(mockRepo)

	mockRepo.EXPECT().GetByEmail("nonexistent@example.com").Return(nil, nil)

	token, err := authService.Login("nonexistent@example.com", "password123")
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Equal(t, "user not found", err.Error())
}

func TestValidateToken_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockUserRepository(ctrl)
	authService := NewAuthService(mockRepo)

	// Create a valid token
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testUser := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Password: string(hashedPassword),
		Role:     "admin",
	}

	mockRepo.EXPECT().GetByEmail("test@example.com").Return(testUser, nil)

	token, err := authService.Login("test@example.com", "password123")
	assert.NoError(t, err)

	// Now validate the token
	claims, err := authService.ValidateToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "1", claims.UserID)
	assert.Equal(t, models.UserRole("admin"), claims.Role)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockUserRepository(ctrl)
	authService := NewAuthService(mockRepo)

	claims, err := authService.ValidateToken("invalid.token.here")
	assert.Error(t, err)
	assert.Nil(t, claims)
}
