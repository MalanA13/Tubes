package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	// Import Handler dan Logika Notification dari folder internal kelompok
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

	r := mux.NewRouter()

	// Inisialisasi Service Notification (dari internal/notification)
	notifRepo := notification.NewNotificationRepository(db)
	notifService := notification.NewNotificationService(notifRepo)

	// Daftarkan Route ke handler kelompok
	r.HandleFunc("/notify", handler.HandleSendNotificationHTTP(*notifService)).Methods("POST")

	log.Println("Notification Service running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
