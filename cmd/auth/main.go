package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux" // Pastikan library router sesuai yang kelompokmu pakai
	models "github.com/tubes-cc/logistics/domain"
	"github.com/tubes-cc/logistics/internal/auth"
	"github.com/tubes-cc/logistics/internal/config"
	"github.com/tubes-cc/logistics/internal/handler"
	"github.com/tubes-cc/logistics/internal/middleware"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	log.Println("Starting Auth-User Service...")

	cfg := config.LoadAuthConfig()

	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
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

	// Health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP","service":"auth"}`))
	}).Methods("GET")

	port := ":" + cfg.Port

	srv := &http.Server{
		Addr:         port,
		Handler:      middleware.CORS(r),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Auth Service is running on port %s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Failed to start server: %v", err)
		}
	}()

	<-ctx.Done()
	stop()
	log.Println("Shutting down gracefully, press Ctrl+C again to force")

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(timeoutCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
