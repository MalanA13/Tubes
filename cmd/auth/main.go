package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux" // Pastikan library router sesuai yang kelompokmu pakai
	models "github.com/tubes-cc/logistics/domain"
	"github.com/tubes-cc/logistics/internal/auth"
	"github.com/tubes-cc/logistics/internal/handler"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	log.Println("Starting Auth-User Service...")

	db, err := gorm.Open(sqlite.Open("auth.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to auth database: %v", err)
	}

	// Auto-migrate User model
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("Failed to auto-migrate User table: %v", err)
	}

	r := mux.NewRouter()

	// 2. Inisialisasi Repository dan Service Auth (dari internal/auth)
	// Sesuaikan dengan nama database/mock yang kamu pakai kemarin
	authRepo := auth.NewUserRepository(db)
	authService := auth.NewAuthService(authRepo)

	// 3. Daftarkan Route menggunakan Handler yang sudah dipindah ke internal/handler
	// Pastikan fungsi di auth_handler.go kamu menggunakan package 'handler'
	r.HandleFunc("/login", handler.HandleLogin(*authService)).Methods("POST")
	r.HandleFunc("/register", handler.HandleRegister(*authService)).Methods("POST")
	r.HandleFunc("/auth/validate", handler.HandleValidateToken(*authService)).Methods("POST")

	log.Println("Auth Service running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r)) // Port internal container tetap 8080
}
