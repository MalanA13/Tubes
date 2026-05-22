package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	// Import Handler dan Logika Tracking dari folder internal kelompok

	"github.com/tubes-cc/logistics/internal/handler"
	"github.com/tubes-cc/logistics/internal/tracking"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	log.Println("Starting Tracking Service...")

	db, err := gorm.Open(sqlite.Open("tracking.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to tracking database: %v", err)
	}

	r := mux.NewRouter()

	// Inisialisasi Service Tracking (dari internal/tracking)
	trackingRepo := tracking.NewTrackingRepository(db)
	trackingService := tracking.NewTrackingService(trackingRepo)

	// Daftarkan Route ke handler kelompok
	r.HandleFunc("/track", handler.HandleSendTrackingHTTP(*trackingService)).Methods("POST")

	port := ":" + getEnv("PORT", "8083") // Tracking service uses port 8083
	log.Printf("Tracking Service is running on port %s", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
