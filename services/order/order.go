package order

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/metric"
)

type Service struct {
	Handler *Handler
	Store   OrderStore
	Mux     *http.ServeMux
}

func NewService(productServiceURL, inventoryServiceURL string, httpClient *http.Client, pool *pgxpool.Pool, ordersCounter metric.Int64Counter) *Service {
	var store OrderStore
	if pool != nil {
		store = NewPGStore(pool)
	} else {
		store = NewMemoryStore()
	}
	client := NewProductClient(productServiceURL, httpClient)
	inventoryClient := NewInventoryClient(inventoryServiceURL, httpClient)
	handler := NewHandler(store, client, inventoryClient, ordersCounter)
	mux := http.NewServeMux()
	RegisterRoutes(mux, handler)

	return &Service{
		Handler: handler,
		Store:   store,
		Mux:     mux,
	}
}
