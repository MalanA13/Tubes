package auth

import (
	"context"
	"strconv"

	models "github.com/tubes-cc/logistics/domain"
	"gorm.io/gorm"
)

// userRepositoryImpl adalah struktur yang akan melakukan query ke database
type userRepositoryImpl struct {
	db *gorm.DB
}

// NewUserRepository adalah fungsi untuk membuat instance repository baru
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepositoryImpl{db: db}
}

// GetByEmail mencari data user asli di database berdasarkan email
func (r *userRepositoryImpl) GetByEmail(email string) (*models.User, error) {
	var user models.User
	// Query GORM: SELECT * FROM users WHERE email = 'email' LIMIT 1;
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

// GetUserByID mencari user berdasarkan ID (string dikonversi ke uint)
func (r *userRepositoryImpl) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	// Konversi string ID ke uint
	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		return nil, gorm.ErrRecordNotFound
	}

	var user models.User
	// Query GORM: SELECT * FROM users WHERE id = ? LIMIT 1;
	err = r.db.WithContext(ctx).Where("id = ?", uint(id)).First(&user).Error
	return &user, err
}

// Create menyimpan data user baru ke database
func (r *userRepositoryImpl) Create(user *models.User) error {
	// Query GORM: INSERT INTO users ...
	return r.db.Create(user).Error
}
