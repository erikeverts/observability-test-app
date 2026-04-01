package inventory

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	Handler *Handler
	Store   InventoryStore
	Mux     *http.ServeMux
}

func NewService(dataDir string, pool *pgxpool.Pool) (*Service, error) {
	var store InventoryStore
	var err error
	if pool != nil {
		store, err = NewPGStore(pool, dataDir)
	} else {
		store, err = NewMemoryStore(dataDir)
	}
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
