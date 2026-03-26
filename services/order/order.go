package order

import "net/http"

type Service struct {
	Handler *Handler
	Store   *Store
	Mux     *http.ServeMux
}

func NewService(productServiceURL string, httpClient *http.Client) *Service {
	store := NewStore()
	client := NewProductClient(productServiceURL, httpClient)
	handler := NewHandler(store, client)
	mux := http.NewServeMux()
	RegisterRoutes(mux, handler)

	return &Service{
		Handler: handler,
		Store:   store,
		Mux:     mux,
	}
}
