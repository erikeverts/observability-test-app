package gateway

import "net/http"

type Service struct {
	Handler *Handler
	Mux     *http.ServeMux
}

func NewService(productServiceURL, orderServiceURL, inventoryServiceURL string, httpClient *http.Client) *Service {
	handler := NewHandler(productServiceURL, orderServiceURL, inventoryServiceURL, httpClient)
	spa := NewSPAHandler()
	mux := http.NewServeMux()
	RegisterRoutes(mux, handler, spa)

	return &Service{
		Handler: handler,
		Mux:     mux,
	}
}
