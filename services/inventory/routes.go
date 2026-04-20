package inventory

import "net/http"

func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("GET /inventory", h.ListStock)
	mux.HandleFunc("GET /inventory/{id}", h.GetStock)
	mux.HandleFunc("POST /inventory/reserve", h.ReserveStock)
	mux.HandleFunc("POST /inventory/restock", h.Restock)
	mux.HandleFunc("GET /inventory/disk-usage", h.DiskUsage)
}
