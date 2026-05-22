package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	// Import Handler dan Logika Pricing dari folder internal kelompok

	"github.com/tubes-cc/logistics/internal/handler"
	"github.com/tubes-cc/logistics/internal/pricing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	log.Println("Starting Pricing Service...")

	db, err := gorm.Open(sqlite.Open("pricing.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to pricing database: %v", err)
	}

	r := mux.NewRouter()

	// Inisialisasi Service Pricing (dari internal/pricing)
	pricingRepo := pricing.NewPricingRepository(db)
	pricingService := pricing.NewPricingService(pricingRepo)

	// Daftarkan Route ke handler kelompok
	r.HandleFunc("/pricing", handler.HandleSendPricingHTTP(*pricingService)).Methods("POST")

	port := ":" + getEnv("PORT", "8082") // Pricing service uses port 8082
	log.Printf("Pricing Service is running on port %s", port)
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
