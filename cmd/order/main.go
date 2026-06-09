package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	// Import Handler dan Logika Pricing dari folder internal kelompok

	"github.com/tubes-cc/logistics/client"
	"github.com/tubes-cc/logistics/internal/handler"
	"github.com/tubes-cc/logistics/internal/order"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	log.Println("Starting Ordering Service...")

	// 1. Setup Database & Repository
	db, err := gorm.Open(sqlite.Open("order.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	orderRepo, err := order.NewOrderRepository(db)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}

	// 2. Setup Clients (Inter-service communication)
	trackingURL := getEnv("TRACKING_SVC_URL", "http://tracking-service:8080")
	pricingURL := getEnv("PRICING_SVC_URL", "http://pricing-service:8080")

	trackingClient := client.NewHTTPTrackingClient(trackingURL)
	pricingClient := client.NewHTTPPricingClient(pricingURL)

	// 3. Setup Service
	orderService := order.NewService(orderRepo, trackingClient, pricingClient)

	// 4. Setup Routes
	router := mux.NewRouter()

	// 5. Register the endpoints
	router.HandleFunc("/order", handler.HandleOrderHTTP(*orderService)).Methods("POST")
	router.HandleFunc("/orders/{resiID}/validate", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		resiID := vars["resiID"]
		if resiID == "" {
			http.Error(w, "resiID is required", http.StatusBadRequest)
			return
		}

		exists, err := orderRepo.ValidateResi(resiID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !exists {
			http.NotFound(w, r)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"valid"}`))
	}).Methods("GET")

	// 6. Start Server
	port := ":" + getEnv("PORT", "8081") // Order service uses port 8081
	log.Printf("Order Service is running on port %s", port)
	if err := http.ListenAndServe(port, router); err != nil {
	log.Fatalf("Failed to start server: %v", err)
}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
