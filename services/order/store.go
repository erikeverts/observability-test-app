package order

import (
	"context"
	"fmt"
	"sync"

	"github.com/erikeverts/observability-test-app/internal/model"
)

type MemoryStore struct {
	mu     sync.RWMutex
	orders map[string]*model.Order
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		orders: make(map[string]*model.Order),
	}
}

func (s *MemoryStore) Save(_ context.Context, order *model.Order) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders[order.ID] = order
	return nil
}

func (s *MemoryStore) Get(_ context.Context, id string) (*model.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.orders[id]
	if !ok {
		return nil, fmt.Errorf("order %s not found", id)
	}
	cp := *o
	return &cp, nil
}

func (s *MemoryStore) List(_ context.Context) ([]model.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]model.Order, 0, len(s.orders))
	for _, o := range s.orders {
		result = append(result, *o)
	}
	return result, nil
}
