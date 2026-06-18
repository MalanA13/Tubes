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

	"github.com/tubes-cc/logistics/domain"
	pginfra "github.com/tubes-cc/logistics/infrastructure/postgres"
	"github.com/tubes-cc/logistics/internal/config"
	"github.com/tubes-cc/logistics/internal/handler"
	"github.com/tubes-cc/logistics/internal/logger"
	"github.com/tubes-cc/logistics/internal/middleware"
	"github.com/tubes-cc/logistics/internal/pricing"

	"go.uber.org/zap"
)

func main() {
	loggerInstance, err := logger.InitLogger("pricing")
	if err != nil {
		zap.L().Fatal("Failed to initialize logger", zap.Error(err))
	}
	defer loggerInstance.Sync()

	zap.L().Info("Starting Pricing Service...")

	cfg := config.LoadPricingConfig()

	db, err := pginfra.Open()
	if err != nil {
		zap.L().Fatal("Failed to connect to pricing database", zap.Error(err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		zap.L().Fatal("Failed to get raw database connection", zap.Error(err))
	}
	defer sqlDB.Close()

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
				zap.L().Error("Failed to seed tariff", zap.String("origin", t.Origin), zap.String("destination", t.Destination), zap.String("service_type", string(t.ServiceType)), zap.Error(err))
			}
		}
	}

	r := mux.NewRouter()

	pricingRepo := pricing.NewPricingRepository(db)
	pricingService := pricing.NewPricingService(pricingRepo)

	r.HandleFunc("/pricing", handler.HandleSendPricingHTTP(*pricingService)).Methods("POST")

	// Health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := sqlDB.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"DOWN","service":"pricing","reason":"database ping failed"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP","service":"pricing"}`))
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
		zap.L().Info("Pricing Service is running", zap.String("port", port))
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