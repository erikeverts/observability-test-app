package order

import (
	"fmt"
	"sync"

	"github.com/erikeverts/observability-test-app/internal/model"
)

type Store struct {
	mu     sync.RWMutex
	orders map[string]*model.Order
}

func NewStore() *Store {
	return &Store{
		orders: make(map[string]*model.Order),
	}
}

func (s *Store) Save(order *model.Order) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders[order.ID] = order
}

func (s *Store) Get(id string) (*model.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.orders[id]
	if !ok {
		return nil, fmt.Errorf("order %s not found", id)
	}
	cp := *o
	return &cp, nil
}

func (s *Store) List() []model.Order {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]model.Order, 0, len(s.orders))
	for _, o := range s.orders {
		result = append(result, *o)
	}
	return result
}
