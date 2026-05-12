package main

import (
	"log"
	"net/http"
	"os"

	"github.com/tubes-cc/logistics/client"
	sqliterepo "github.com/tubes-cc/logistics/infrastructure/sqlite"
	"github.com/tubes-cc/logistics/internal/handler"
	"github.com/tubes-cc/logistics/internal/hub"
	"github.com/tubes-cc/logistics/internal/middleware"
)

func main() {
	log.Println("Starting Hub Service...")

	// 1. Setup Database
	db, err := sqliterepo.Open("hub.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	repo, err := sqliterepo.NewShipmentRepository(db)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}

	// 2. Setup Clients (Inter-service communication)
	trackingURL := getEnv("TRACKING_SVC_URL", "http://tracking-service:8080")
	orderURL := getEnv("ORDER_SVC_URL", "http://order-service:8080")
	authURL := getEnv("AUTH_SVC_URL", "http://auth-service:8080")

	trackingClient := client.NewHTTPTrackingClient(trackingURL)
	orderClient := client.NewHTTPOrderClient(orderURL)
	authClient := client.NewHTTPAuthClient(authURL)

	// 3. Setup Service
	hubService := hub.NewService(repo, trackingClient, orderClient)

	// 4. Setup Handlers & Middleware
	hubHandler := handler.NewHubHandler(hubService)
	authMiddleware := middleware.NewAuthMiddleware(authClient)

	// 5. Setup Routes
	mux := http.NewServeMux()

	// Protected routes using middleware
	mux.HandleFunc("/hub/scan-in", authMiddleware.Authenticate(
		authMiddleware.RequireRole("admin", hubHandler.ScanIn),
	))

	mux.HandleFunc("/hub/scan-out", authMiddleware.Authenticate(
		authMiddleware.RequireRole("admin", hubHandler.ScanOut),
	))

	// 6. Start Server
	port := ":" + getEnv("PORT", "8084") // Hub service uses port 8084
	log.Printf("Hub Service is running on port %s", port)
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
