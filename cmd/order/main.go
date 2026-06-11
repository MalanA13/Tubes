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

	"github.com/gorilla/mux"
	// Import Handler dan Logika Pricing dari folder internal kelompok

	"github.com/tubes-cc/logistics/client"
	"github.com/tubes-cc/logistics/internal/config"
	"github.com/tubes-cc/logistics/internal/handler"
	"github.com/tubes-cc/logistics/internal/middleware"
	"github.com/tubes-cc/logistics/internal/order"
	"github.com/tubes-cc/logistics/internal/response"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	log.Println("Starting Ordering Service...")

	cfg := config.LoadOrderConfig()

	// 1. Setup Database & Repository
	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	orderRepo, err := order.NewOrderRepository(db)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
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
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP","service":"order"}`))
	}).Methods("GET")

	// 6. Start Server
	port := ":" + cfg.Port

	srv := &http.Server{
		Addr:         port,
		Handler:      middleware.CORS(router),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Order Service is running on port %s", port)
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
