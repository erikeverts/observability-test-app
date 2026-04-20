package gateway

import "net/http"

func RegisterRoutes(mux *http.ServeMux, h *Handler, spa http.Handler) {
	// API routes (proxied to backend services)
	mux.HandleFunc("GET /products", h.ProxyProducts)
	mux.HandleFunc("GET /products/{id}", h.ProxyProducts)
	mux.HandleFunc("GET /orders", h.ProxyOrders)
	mux.HandleFunc("GET /orders/{id}", h.ProxyOrders)
	mux.HandleFunc("POST /orders", h.ProxyOrders)
	mux.HandleFunc("GET /inventory", h.ProxyInventory)
	mux.HandleFunc("GET /inventory/{id}", h.ProxyInventory)
	mux.HandleFunc("POST /inventory/reserve", h.ProxyInventory)
	mux.HandleFunc("POST /inventory/restock", h.ProxyInventory)
	mux.HandleFunc("GET /inventory/disk-usage", h.ProxyInventory)

	// SPA frontend (fallback for all other GET requests)
	mux.Handle("GET /", spa)
}
