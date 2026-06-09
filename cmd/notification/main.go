package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	models "github.com/tubes-cc/logistics/domain"
	"github.com/tubes-cc/logistics/internal/handler"
	"github.com/tubes-cc/logistics/internal/notification"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	log.Println("Starting Notification Service...")

	db, err := gorm.Open(sqlite.Open("notification.db"), &gorm.Config{})
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

	log.Println("Notification Service running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
