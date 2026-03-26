package gateway

import "net/http"

type Service struct {
	Handler *Handler
	Mux     *http.ServeMux
}

func NewService(productServiceURL, orderServiceURL, inventoryServiceURL string, httpClient *http.Client) *Service {
	handler := NewHandler(productServiceURL, orderServiceURL, inventoryServiceURL, httpClient)
	mux := http.NewServeMux()
	RegisterRoutes(mux, handler)

	return &Service{
		Handler: handler,
		Mux:     mux,
	}
}
