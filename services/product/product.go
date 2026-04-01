package product

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	Handler *Handler
	Store   ProductStore
	Mux     *http.ServeMux
}

func NewService(pool *pgxpool.Pool) *Service {
	var store ProductStore
	if pool != nil {
		store = NewPGStore(pool)
	} else {
		store = NewMemoryStore()
	}
	handler := NewHandler(store)
	mux := http.NewServeMux()
	RegisterRoutes(mux, handler)

	return &Service{
		Handler: handler,
		Store:   store,
		Mux:     mux,
	}
}
