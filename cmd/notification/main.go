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
	models "github.com/tubes-cc/logistics/domain"
	"github.com/tubes-cc/logistics/internal/config"
	"github.com/tubes-cc/logistics/internal/handler"
	"github.com/tubes-cc/logistics/internal/middleware"
	"github.com/tubes-cc/logistics/internal/notification"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	log.Println("Starting Notification Service...")

	cfg := config.LoadNotificationConfig()

	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to notification database: %v", err)
	}

	// Auto-migrate: creates the notifications table if it doesn't exist
	if err := db.AutoMigrate(&models.Notification{}); err != nil {
		log.Fatalf("Failed to auto-migrate notification table: %v", err)
	}

	// Initialize Repository and Service
	notifRepo := notification.NewNotificationRepository(db)
	notifService := notification.NewNotificationService(notifRepo)

	r := mux.NewRouter()

	// Register the POST /notify route with the real handler
	r.HandleFunc("/notify", handler.HandleSendNotificationHTTP(*notifService)).Methods("POST")

	// Health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP","service":"notification"}`))
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
		log.Printf("Notification Service is running on port %s", port)
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
