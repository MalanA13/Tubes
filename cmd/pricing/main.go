package main

import (
	"log"
	"net/http"

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

	log.Println("Pricing Service running on port 8082...")
	log.Fatal(http.ListenAndServe(":8082", r))
}
