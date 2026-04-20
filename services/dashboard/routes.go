package dashboard

import "net/http"

func RegisterRoutes(mux *http.ServeMux, h *Handler, staticFS http.Handler) {
	mux.HandleFunc("GET /api/services", h.ListServices)
	mux.HandleFunc("GET /api/services/{name}/chaos", h.GetServiceChaos)
	mux.HandleFunc("PUT /api/services/{name}/chaos", h.SetServiceChaos)
	mux.HandleFunc("POST /api/services/{name}/chaos/clear-disk", h.ClearServiceDisk)
	mux.Handle("/", staticFS)
}
