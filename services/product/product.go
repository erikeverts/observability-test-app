package product

import "net/http"

type Service struct {
	Handler *Handler
	Store   *Store
	Mux     *http.ServeMux
}

func NewService() *Service {
	store := NewStore()
	handler := NewHandler(store)
	mux := http.NewServeMux()
	RegisterRoutes(mux, handler)

	return &Service{
		Handler: handler,
		Store:   store,
		Mux:     mux,
	}
}
