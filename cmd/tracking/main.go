package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	// Import Handler dan Logika Tracking dari folder internal kelompok

	"github.com/tubes-cc/logistics/client"
	"github.com/tubes-cc/logistics/internal/config"
	"github.com/tubes-cc/logistics/internal/handler"
	"github.com/tubes-cc/logistics/internal/logger"
	"github.com/tubes-cc/logistics/internal/middleware"
	"github.com/tubes-cc/logistics/internal/tracking"

	"go.uber.org/zap"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	loggerInstance, err := logger.InitLogger("tracking")
	if err != nil {
		zap.L().Fatal("Failed to initialize logger", zap.Error(err))
	}
	defer loggerInstance.Sync()

	zap.L().Info("Starting Tracking Service...")

	cfg := config.LoadTrackingConfig()

	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		zap.L().Fatal("Failed to connect to tracking database", zap.Error(err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		zap.L().Fatal("Failed to get raw database connection", zap.Error(err))
	}
	defer sqlDB.Close()

	// Auto-migrate Tracking models
	if err := db.AutoMigrate(&tracking.TrackingEventModel{}, &tracking.ShipmentModel{}); err != nil {
		zap.L().Fatal("Failed to auto-migrate tracking tables", zap.Error(err))
	}

	r := mux.NewRouter()

	// Inisialisasi Service Tracking (dari internal/tracking)
	trackingRepo := tracking.NewTrackingRepository(db)
	trackingService := tracking.NewTrackingService(trackingRepo)

	// Setup Auth Client and Middleware
	authClient := client.NewHTTPAuthClient(cfg.AuthSvcURL)
	authMiddleware := middleware.NewAuthMiddleware(authClient)

	// Daftarkan Route ke handler kelompok
	// POST endpoints require admin role
	r.Handle("/track", authMiddleware.Authenticate(
		authMiddleware.RequireRole("admin", handler.HandleSendTrackingHTTP(*trackingService)),
	)).Methods("POST")
	r.Handle("/events", authMiddleware.Authenticate(
		authMiddleware.RequireRole("admin", handler.HandleSendTrackingHTTP(*trackingService)),
	)).Methods("POST")

	// GET endpoints require authentication only
	r.Handle("/tracking/{resiID}/status", authMiddleware.Authenticate(
		handler.HandleGetCurrentStatusHTTP(*trackingService),
	)).Methods("GET")
	r.Handle("/tracking/{resiID}", authMiddleware.Authenticate(
		handler.HandleGetTrackingHTTP(*trackingService),
	)).Methods("GET")

	// Health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := sqlDB.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"DOWN","service":"tracking","reason":"database ping failed"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP","service":"tracking"}`))
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
		zap.L().Info("Tracking Service is running", zap.String("port", port))
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
