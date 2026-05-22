package main

import (
	"log"
	"net/http"

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

	log.Println("Tracking Service running on port 8083...")
	log.Fatal(http.ListenAndServe(":8083", r))
}
