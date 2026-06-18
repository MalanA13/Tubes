package functional

import (
	"strings"
	"testing"

	models "github.com/tubes-cc/logistics/domain"
	"github.com/tubes-cc/logistics/internal/auth"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates a fresh in-memory SQLite database for each test.
// Using ":memory:" ensures complete isolation — no stale test.db file,
// no UNIQUE constraint collisions between runs.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create in-memory SQLite database: %v", err)
	}
	// Auto-migrate creates the users table from scratch every time
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("Failed to auto-migrate: %v", err)
	}
	return db
}

func TestLogin_Functional(t *testing.T) {
	// 1. Setup: fresh in-memory database (no leftover data, no UNIQUE conflicts)
	db := setupTestDB(t)

	// 2. Initialize real repository and service
	repo := auth.NewUserRepository(db)
	authService := auth.NewAuthService(repo)

	// 3. Register user via AuthService.Register() — this hashes the password
	//    with bcrypt internally. We must NOT insert plain-text passwords via
	//    repo.Create() because Login() uses bcrypt.CompareHashAndPassword().
	_, err := authService.Register("Budi Kurir", "budi@ekspedisi.com", "password_aman")
	assert.NoError(t, err, "Register should succeed")

	// 4. Login with the same credentials
	token, err := authService.Login("budi@ekspedisi.com", "password_aman")

	// 5. Assert: no error, token is a valid JWT (3 dot-separated parts)
	assert.NoError(t, err, "Login should succeed with correct credentials")
	assert.NotEmpty(t, token, "Token should not be empty")

	parts := strings.Split(token, ".")
	assert.Equal(t, 3, len(parts), "JWT token must have 3 dot-separated parts (header.payload.signature)")

	// 6. Validate the token using the service's own ValidateToken method
	claims, err := authService.ValidateToken(token)
	assert.NoError(t, err, "Token validation should succeed")
	assert.NotNil(t, claims, "Claims should not be nil")
	assert.NotEmpty(t, claims.UserID, "UserID claim should be set")
}

func TestLogin_WrongPassword_Functional(t *testing.T) {
	db := setupTestDB(t)
	repo := auth.NewUserRepository(db)
	authService := auth.NewAuthService(repo)

	// Register a user
	_, err := authService.Register("Budi Kurir", "budi@ekspedisi.com", "password_aman")
	assert.NoError(t, err)

	// Attempt login with wrong password
	token, err := authService.Login("budi@ekspedisi.com", "wrong_password")
	assert.Error(t, err, "Login should fail with wrong password")
	assert.Empty(t, token, "Token should be empty on failed login")
	assert.Equal(t, "incorrect password", err.Error())
}

func TestLogin_UserNotFound_Functional(t *testing.T) {
	db := setupTestDB(t)
	repo := auth.NewUserRepository(db)
	authService := auth.NewAuthService(repo)

	// Attempt login without registering
	token, err := authService.Login("nobody@ekspedisi.com", "password_aman")
	assert.Error(t, err, "Login should fail for non-existent user")
	assert.Empty(t, token)
	assert.Equal(t, "user not found", err.Error())
}

func TestRegister_DuplicateEmail_Functional(t *testing.T) {
	db := setupTestDB(t)
	repo := auth.NewUserRepository(db)
	authService := auth.NewAuthService(repo)

	// Register first user
	_, err := authService.Register("Budi Kurir", "budi@ekspedisi.com", "password_aman")
	assert.NoError(t, err)

	// Attempt to register again with the same email
	_, err = authService.Register("Budi Lain", "budi@ekspedisi.com", "password_lain")
	assert.Error(t, err, "Should fail on duplicate email")
	assert.Equal(t, "user already exists", err.Error())
}