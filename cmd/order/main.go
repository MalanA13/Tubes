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
	// Import Handler dan Logika Pricing dari folder internal kelompok

	"github.com/tubes-cc/logistics/client"
	"github.com/tubes-cc/logistics/internal/config"
	"github.com/tubes-cc/logistics/internal/handler"
	"github.com/tubes-cc/logistics/internal/logger"
	"github.com/tubes-cc/logistics/internal/middleware"
	"github.com/tubes-cc/logistics/internal/order"
	"github.com/tubes-cc/logistics/internal/response"

	"go.uber.org/zap"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	loggerInstance, err := logger.InitLogger("order")
	if err != nil {
		zap.L().Fatal("Failed to initialize logger", zap.Error(err))
	}
	defer loggerInstance.Sync()

	zap.L().Info("Starting Ordering Service...")

	cfg := config.LoadOrderConfig()

	// 1. Setup Database & Repository
	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		zap.L().Fatal("Failed to open database", zap.Error(err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		zap.L().Fatal("Failed to get raw database connection", zap.Error(err))
	}
	defer sqlDB.Close()

	orderRepo, err := order.NewOrderRepository(db)
	if err != nil {
		zap.L().Fatal("Failed to initialize repository", zap.Error(err))
	}

	// 2. Setup Clients (Inter-service communication)
	trackingClient := client.NewHTTPTrackingClient(cfg.TrackingSvcURL)
	pricingClient := client.NewHTTPPricingClient(cfg.PricingSvcURL)

	// 3. Setup Service
	orderService := order.NewService(orderRepo, trackingClient, pricingClient)

	// 4. Setup Routes
	router := mux.NewRouter()

	// 5. Register the endpoints
	router.HandleFunc("/order", handler.HandleOrderHTTP(*orderService)).Methods("POST")
	router.HandleFunc("/orders/{resiID}/validate", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		resiID := vars["resiID"]
		if resiID == "" {
			response.BadRequest(w, "resiID is required")
			return
		}

		exists, err := orderRepo.ValidateResi(resiID)
		if err != nil {
			response.InternalServerError(w, err.Error())
			return
		}
		if !exists {
			response.NotFound(w, "resi not found")
			return
		}

		response.OK(w, map[string]string{"status": "valid"})
	}).Methods("GET")

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := sqlDB.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"DOWN","service":"order","reason":"database ping failed"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP","service":"order"}`))
	}).Methods("GET")

	// 6. Start Server
	port := ":" + cfg.Port

	srv := &http.Server{
		Addr:         port,
		Handler:      middleware.CORS(middleware.RequestIDMiddleware(middleware.LoggingMiddleware(middleware.RecoveryMiddleware(router)))),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		zap.L().Info("Order Service is running", zap.String("port", port))
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
