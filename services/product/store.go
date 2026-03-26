package product

import (
	"fmt"
	"sync"

	"github.com/erikeverts/observability-test-app/internal/model"
)

type Store struct {
	mu       sync.RWMutex
	products map[string]*model.Product
}

func NewStore() *Store {
	s := &Store{
		products: make(map[string]*model.Product),
	}
	s.seed()
	return s
}

func (s *Store) seed() {
	products := []model.Product{
		{ID: "prod-1", Name: "Mechanical Keyboard", Description: "Cherry MX Blue switches", Price: 149.99, Stock: 50},
		{ID: "prod-2", Name: "Wireless Mouse", Description: "Ergonomic design, 2.4GHz", Price: 49.99, Stock: 120},
		{ID: "prod-3", Name: "USB-C Hub", Description: "7-in-1 multiport adapter", Price: 39.99, Stock: 200},
		{ID: "prod-4", Name: "Monitor Stand", Description: "Adjustable aluminum stand", Price: 79.99, Stock: 35},
		{ID: "prod-5", Name: "Webcam HD", Description: "1080p with autofocus", Price: 69.99, Stock: 80},
	}
	for i := range products {
		s.products[products[i].ID] = &products[i]
	}
}

func (s *Store) List() []model.Product {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]model.Product, 0, len(s.products))
	for _, p := range s.products {
		result = append(result, *p)
	}
	return result
}

func (s *Store) Get(id string) (*model.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.products[id]
	if !ok {
		return nil, fmt.Errorf("product %s not found", id)
	}
	cp := *p
	return &cp, nil
}
