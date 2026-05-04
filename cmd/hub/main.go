package main

import (
	"log"
	"net/http"

	"github.com/tubes-cc/logistics/client"
	"github.com/tubes-cc/logistics/internal/handler"
	"github.com/tubes-cc/logistics/internal/hub"
	"github.com/tubes-cc/logistics/internal/middleware"
	sqliterepo "github.com/tubes-cc/logistics/infrastructure/sqlite"
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
	// In a real scenario, these URLs would come from environment variables
	trackingClient := client.NewHTTPTrackingClient("http://tracking-service:8080")
	orderClient := client.NewHTTPOrderClient("http://order-service:8080")
	authClient := client.NewHTTPAuthClient("http://auth-service:8080")

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
	port := ":8081"
	log.Printf("Hub Service is running on port %s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
