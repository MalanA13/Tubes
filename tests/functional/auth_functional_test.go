package functional

import (
	"testing"
    models "github.com/tubes-cc/logistics/domain"           // Mengarah ke domain kelompok
    "github.com/tubes-cc/logistics/internal/auth"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite" // Import driver SQLite
	"gorm.io/gorm"
)

func TestLogin_Functional(t *testing.T) {
	// 1. Setup Database (File 'test.db' akan dibuat otomatis di foldermu)
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Gagal membuat database SQLite: %v", err)
	}

	// 2. Auto Migrate: GORM akan otomatis membuat tabel 'users' berdasarkan struct model
	db.AutoMigrate(&models.User{})

	// Bersihkan data lama (agar test bisa dijalankan berkali-kali tanpa error duplicate)
	db.Exec("DELETE FROM users")

	// 3. Inisialisasi Repository dan Service yang ASLI
	repo := auth.NewUserRepository(db)
	authService := auth.NewAuthService(repo)

	// 4. Masukkan data nyata ke Database SQLite
	newUser := &models.User{
		FullName: "Budi Kurir",
		Email:    "budi@ekspedisi.com",
		Password: "password_aman",
	}
	repo.Create(newUser)

	// 5. Jalankan Tes: Coba login dengan data yang barusan kita simpan
	token, err := authService.Login("budi@ekspedisi.com", "password_aman")

	// 6. Assert: Pastikan tidak ada error dan token keluar
	assert.NoError(t, err)
	assert.Equal(t, "TOKEN_JWT_DUMMY", token)
}