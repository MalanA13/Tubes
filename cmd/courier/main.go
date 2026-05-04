package main

import (
	"log"
	"net/http"
	"os"

	"github.com/tubes-cc/logistics/client"
	"github.com/tubes-cc/logistics/internal/courier"
	"github.com/tubes-cc/logistics/internal/handler"
	"github.com/tubes-cc/logistics/internal/middleware"
	sqliterepo "github.com/tubes-cc/logistics/infrastructure/sqlite"
)

func main() {
	log.Println("Starting Courier Service...")

	// 1. Setup Database
	db, err := sqliterepo.Open("courier.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	repo, err := sqliterepo.NewShipmentRepository(db)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}

	// 2. Setup Clients
	trackingURL := getEnv("TRACKING_SVC_URL", "http://tracking-service:8080")
	orderURL := getEnv("ORDER_SVC_URL", "http://order-service:8080")
	authURL := getEnv("AUTH_SVC_URL", "http://auth-service:8080")

	trackingClient := client.NewHTTPTrackingClient(trackingURL)
	orderClient := client.NewHTTPOrderClient(orderURL)
	authClient := client.NewHTTPAuthClient(authURL)

	// 3. Setup Service
	courierService := courier.NewService(repo, trackingClient, orderClient)

	// 4. Setup Handlers & Middleware
	courierHandler := handler.NewCourierHandler(courierService)
	authMiddleware := middleware.NewAuthMiddleware(authClient)

	// 5. Setup Routes
	mux := http.NewServeMux()

	mux.HandleFunc("/courier/assign", authMiddleware.Authenticate(
		authMiddleware.RequireRole("admin", courierHandler.AssignCourier),
	))
	
	mux.HandleFunc("/courier/delivery-status", authMiddleware.Authenticate(
		authMiddleware.RequireRole("courier", courierHandler.UpdateDeliveryStatus),
	))

	// 6. Start Server
	port := ":" + getEnv("PORT", "8082")
	log.Printf("Courier Service is running on port %s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
