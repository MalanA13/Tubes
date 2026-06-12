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

	"github.com/tubes-cc/logistics/client"
	sqliterepo "github.com/tubes-cc/logistics/infrastructure/sqlite"
	"github.com/tubes-cc/logistics/internal/logger"
	"github.com/tubes-cc/logistics/internal/config"
	"github.com/tubes-cc/logistics/internal/handler"
	"github.com/tubes-cc/logistics/internal/hub"
	"github.com/tubes-cc/logistics/internal/middleware"

	"go.uber.org/zap"
	"github.com/tubes-cc/logistics/internal/response"
)

func main() {
	loggerInstance, err := logger.InitLogger("hub")
	if err != nil {
		zap.L().Fatal("Failed to initialize logger", zap.Error(err))
	}
	defer loggerInstance.Sync()

	zap.L().Info("Starting Hub Service...")

	cfg := config.LoadHubConfig()

	// 1. Setup Database
	db, err := sqliterepo.Open(cfg.DBPath)
	if err != nil {
		zap.L().Fatal("Failed to open database", zap.Error(err))
	}
	defer db.Close()

	repo, err := sqliterepo.NewShipmentRepository(db)
	if err != nil {
		zap.L().Fatal("Failed to initialize repository", zap.Error(err))
	}

	// 2. Setup Clients (Inter-service communication)
	trackingClient := client.NewHTTPTrackingClient(cfg.TrackingSvcURL)
	orderClient := client.NewHTTPOrderClient(cfg.OrderSvcURL)
	authClient := client.NewHTTPAuthClient(cfg.AuthSvcURL)

	// 3. Setup Service
	hubService := hub.NewService(repo, trackingClient, orderClient)

	// 4. Setup Handlers & Middleware
	hubHandler := handler.NewHubHandler(hubService)
	authMiddleware := middleware.NewAuthMiddleware(authClient)

	// 5. Setup Routes
	mux := http.NewServeMux()

	// Protected routes using middleware
	mux.HandleFunc("/hub/scan-in", authMiddleware.Authenticate(
		authMiddleware.RequireRole("admin", hubHandler.ScanIn),
	))

	mux.HandleFunc("/hub/scan-out", authMiddleware.Authenticate(
		authMiddleware.RequireRole("admin", hubHandler.ScanOut),
	))

	// Health check endpoint (no auth required)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			response.MethodNotAllowed(w, "Method not allowed")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP","service":"hub"}`))
	})

	// 6. Start Server
	port := ":" + cfg.Port // Hub service uses port 8084

	srv := &http.Server{
		Addr:         port,
		Handler:      middleware.CORS(middleware.RequestIDMiddleware(middleware.LoggingMiddleware(mux))),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		zap.L().Info(fmt.Sprintf("Hub Service is running on port %s", port))
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
