package main

import (
	"log"
	"net/http"
	"os"

	"github.com/erikeverts/observability-test-app/services/dashboard"
)

func main() {
	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "8083"
	}

	gatewayURL := os.Getenv("GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "http://localhost:8080"
	}
	productURL := os.Getenv("PRODUCT_SERVICE_URL")
	if productURL == "" {
		productURL = "http://localhost:8081"
	}
	orderURL := os.Getenv("ORDER_SERVICE_URL")
	if orderURL == "" {
		orderURL = "http://localhost:8082"
	}
	inventoryURL := os.Getenv("INVENTORY_SERVICE_URL")
	if inventoryURL == "" {
		inventoryURL = "http://localhost:8084"
	}

	svc := dashboard.NewService(gatewayURL, productURL, orderURL, inventoryURL)

	addr := ":" + port
	log.Printf("starting chaos dashboard on %s", addr)
	if err := http.ListenAndServe(addr, svc.Mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
