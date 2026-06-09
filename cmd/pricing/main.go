package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	// Import Handler dan Logika Pricing dari folder internal kelompok

	"github.com/tubes-cc/logistics/domain"
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

	// Auto-migrate Tariff model
	if err := db.AutoMigrate(&pricing.Tariff{}); err != nil {
		log.Fatalf("Failed to auto-migrate Tariff table: %v", err)
	}

	// Seed default tariff data if empty
	var count int64
	if err := db.Model(&pricing.Tariff{}).Count(&count).Error; err == nil && count == 0 {
		log.Println("Seeding default tariffs...")
		defaultTariffs := []pricing.Tariff{
			{Origin: "Jakarta", Destination: "Bandung", ServiceType: domain.ServiceExpress, BaseCost: 10000.0},
			{Origin: "Jakarta", Destination: "Jakarta", ServiceType: domain.ServiceSameday, BaseCost: 2000.0},
			{Origin: "Jakarta", Destination: "Jakarta", ServiceType: domain.ServiceRegular, BaseCost: 5000.0},
			{Origin: "Jakarta", Destination: "Surabaya", ServiceType: domain.ServiceRegular, BaseCost: 15000.0},
			{Origin: "Jakarta", Destination: "Surabaya", ServiceType: domain.ServiceExpress, BaseCost: 25000.0},
		}
		for _, t := range defaultTariffs {
			if err := db.Create(&t).Error; err != nil {
				log.Printf("Failed to seed tariff %s -> %s (%s): %v", t.Origin, t.Destination, t.ServiceType, err)
			}
		}
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
