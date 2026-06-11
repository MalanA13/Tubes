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

	"github.com/tubes-cc/logistics/domain"
	"github.com/tubes-cc/logistics/internal/config"
	"github.com/tubes-cc/logistics/internal/handler"
	"github.com/tubes-cc/logistics/internal/middleware"
	"github.com/tubes-cc/logistics/internal/pricing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	log.Println("Starting Pricing Service...")

	cfg := config.LoadPricingConfig()

	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
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

	// Health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP","service":"pricing"}`))
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
		log.Printf("Pricing Service is running on port %s", port)
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
