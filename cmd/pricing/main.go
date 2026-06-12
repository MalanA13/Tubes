package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	// Import Handler dan Logika Pricing dari folder internal kelompok

	"github.com/tubes-cc/logistics/domain"
	"github.com/tubes-cc/logistics/internal/logger"
	"github.com/tubes-cc/logistics/internal/config"
	"github.com/tubes-cc/logistics/internal/handler"
	"github.com/tubes-cc/logistics/internal/middleware"
	"github.com/tubes-cc/logistics/internal/pricing"

	"go.uber.org/zap"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	loggerInstance, err := logger.InitLogger("pricing")
	if err != nil {
		zap.L().Fatal("Failed to initialize logger", zap.Error(err))
	}
	defer loggerInstance.Sync()

	zap.L().Info("Starting Pricing Service...")

	cfg := config.LoadPricingConfig()

	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		zap.L().Fatal("Failed to connect to pricing database", zap.Error(err))
	}

	// Auto-migrate Tariff model
	if err := db.AutoMigrate(&pricing.Tariff{}); err != nil {
		zap.L().Fatal("Failed to auto-migrate Tariff table", zap.Error(err))
	}

	// Seed default tariff data if empty
	var count int64
	if err := db.Model(&pricing.Tariff{}).Count(&count).Error; err == nil && count == 0 {
		zap.L().Info("Seeding default tariffs...")
		defaultTariffs := []pricing.Tariff{
			{Origin: "Jakarta", Destination: "Bandung", ServiceType: domain.ServiceExpress, BaseCost: 10000.0},
			{Origin: "Jakarta", Destination: "Jakarta", ServiceType: domain.ServiceSameday, BaseCost: 2000.0},
			{Origin: "Jakarta", Destination: "Jakarta", ServiceType: domain.ServiceRegular, BaseCost: 5000.0},
			{Origin: "Jakarta", Destination: "Surabaya", ServiceType: domain.ServiceRegular, BaseCost: 15000.0},
			{Origin: "Jakarta", Destination: "Surabaya", ServiceType: domain.ServiceExpress, BaseCost: 25000.0},
		}
		for _, t := range defaultTariffs {
			if err := db.Create(&t).Error; err != nil {
				zap.L().Info(fmt.Sprintf("Failed to seed tariff %s -> %s (%s)): %v", t.Origin, t.Destination, t.ServiceType, err))
			}
		}
	}

	r := mux.NewRouter()

	// Inisialisasi Service Pricing (dari internal/pricing)
	pricingRepo := pricing.NewPricingRepository(db)
	pricingService := pricing.NewPricingService(pricingRepo)

	// Daftarkan Route ke handler kelompok
	r.HandleFunc("/pricing", handler.HandleSendPricingHTTP(*pricingService)).Methods("POST")

	// Health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP","service":"pricing"}`))
	}).Methods("GET")

	port := ":" + cfg.Port

	srv := &http.Server{
		Addr:         port,
		Handler:      middleware.CORS(middleware.RequestIDMiddleware(middleware.LoggingMiddleware(r))),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		zap.L().Info(fmt.Sprintf("Pricing Service is running on port %s", port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zap.L().Error("Failed to start server", zap.Error(err))
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
