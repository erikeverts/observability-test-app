package product

import (
	"context"
	"testing"
)

func TestStore_Seeded(t *testing.T) {
	s := NewMemoryStore()
	products, err := s.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(products) != 5 {
		t.Errorf("expected 5 seeded products, got %d", len(products))
	}
}

func TestStore_Get(t *testing.T) {
	s := NewMemoryStore()
	p, err := s.Get(context.Background(), "prod-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID != "prod-1" {
		t.Errorf("expected prod-1, got %s", p.ID)
	}
}

func TestStore_GetNotFound(t *testing.T) {
	s := NewMemoryStore()
	_, err := s.Get(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent product")
	}
}
