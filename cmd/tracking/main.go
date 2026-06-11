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
	// Import Handler dan Logika Tracking dari folder internal kelompok

	"github.com/tubes-cc/logistics/internal/config"
	"github.com/tubes-cc/logistics/internal/handler"
	"github.com/tubes-cc/logistics/internal/middleware"
	"github.com/tubes-cc/logistics/internal/tracking"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	log.Println("Starting Tracking Service...")

	cfg := config.LoadTrackingConfig()

	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to tracking database: %v", err)
	}

	// Auto-migrate Tracking models
	if err := db.AutoMigrate(&tracking.TrackingEventModel{}, &tracking.ShipmentModel{}); err != nil {
		log.Fatalf("Failed to auto-migrate tracking tables: %v", err)
	}

	r := mux.NewRouter()

	// Inisialisasi Service Tracking (dari internal/tracking)
	trackingRepo := tracking.NewTrackingRepository(db)
	trackingService := tracking.NewTrackingService(trackingRepo)

	// Daftarkan Route ke handler kelompok
	r.HandleFunc("/track", handler.HandleSendTrackingHTTP(*trackingService)).Methods("POST")
	r.HandleFunc("/events", handler.HandleSendTrackingHTTP(*trackingService)).Methods("POST")

	// Health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP","service":"tracking"}`))
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
		log.Printf("Tracking Service is running on port %s", port)
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
