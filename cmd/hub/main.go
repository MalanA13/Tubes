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

	"github.com/tubes-cc/logistics/client"
	sqliterepo "github.com/tubes-cc/logistics/infrastructure/sqlite"
	"github.com/tubes-cc/logistics/internal/config"
	"github.com/tubes-cc/logistics/internal/handler"
	"github.com/tubes-cc/logistics/internal/hub"
	"github.com/tubes-cc/logistics/internal/middleware"
	"github.com/tubes-cc/logistics/internal/response"
)

func main() {
	log.Println("Starting Hub Service...")

	cfg := config.LoadHubConfig()

	// 1. Setup Database
	db, err := sqliterepo.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	repo, err := sqliterepo.NewShipmentRepository(db)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
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
		Handler:      middleware.CORS(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Hub Service is running on port %s", port)
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
