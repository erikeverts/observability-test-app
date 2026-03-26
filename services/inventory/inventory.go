package inventory

import "net/http"

type Service struct {
	Handler *Handler
	Store   *Store
	Mux     *http.ServeMux
}

func NewService(dataDir string) (*Service, error) {
	store, err := NewStore(dataDir)
	if err != nil {
		return nil, err
	}
	handler := NewHandler(store)
	mux := http.NewServeMux()
	RegisterRoutes(mux, handler)

	return &Service{
		Handler: handler,
		Store:   store,
		Mux:     mux,
	}, nil
}
