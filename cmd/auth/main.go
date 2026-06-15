package main

import (
	"context"
	"errors"
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
	"github.com/tubes-cc/logistics/internal/logger"
	"github.com/tubes-cc/logistics/internal/middleware"

	"go.uber.org/zap"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	loggerInstance, err := logger.InitLogger("auth")
	if err != nil {
		zap.L().Fatal("Failed to initialize logger", zap.Error(err))
	}
	defer loggerInstance.Sync()

	zap.L().Info("Starting Auth-User Service...")

	cfg := config.LoadAuthConfig()

	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		zap.L().Fatal("Failed to connect to auth database", zap.Error(err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		zap.L().Fatal("Failed to get raw database connection", zap.Error(err))
	}
	defer sqlDB.Close()

	// Auto-migrate User model
	if err := db.AutoMigrate(&models.User{}); err != nil {
		zap.L().Fatal("Failed to auto-migrate User table", zap.Error(err))
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
		if err := sqlDB.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"DOWN","service":"auth","reason":"database ping failed"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP","service":"auth"}`))
	}).Methods("GET")

	port := ":" + cfg.Port

	srv := &http.Server{
		Addr:         port,
		Handler:      middleware.CORS(middleware.RequestIDMiddleware(middleware.LoggingMiddleware(middleware.RecoveryMiddleware(r)))),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		zap.L().Info("Auth Service is running", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zap.L().Fatal("Failed to start server", zap.Error(err))
		}
	}()

	<-ctx.Done()
	stop()
	zap.L().Info("Shutting down gracefully, press Ctrl+C again to force")

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(timeoutCtx); err != nil {
		zap.L().Error("Server forced to shutdown", zap.Error(err))
	}

	zap.L().Info("Server exiting")
}
