package gateway

import "net/http"

func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("GET /products", h.ProxyProducts)
	mux.HandleFunc("GET /products/{id}", h.ProxyProducts)
	mux.HandleFunc("GET /orders", h.ProxyOrders)
	mux.HandleFunc("GET /orders/{id}", h.ProxyOrders)
	mux.HandleFunc("POST /orders", h.ProxyOrders)
	mux.HandleFunc("GET /inventory", h.ProxyInventory)
	mux.HandleFunc("GET /inventory/{id}", h.ProxyInventory)
	mux.HandleFunc("POST /inventory/reserve", h.ProxyInventory)
	mux.HandleFunc("GET /inventory/disk-usage", h.ProxyInventory)
}
