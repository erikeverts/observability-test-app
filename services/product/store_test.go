package product

import "testing"

func TestStore_Seeded(t *testing.T) {
	s := NewStore()
	products := s.List()
	if len(products) != 5 {
		t.Errorf("expected 5 seeded products, got %d", len(products))
	}
}

func TestStore_Get(t *testing.T) {
	s := NewStore()
	p, err := s.Get("prod-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID != "prod-1" {
		t.Errorf("expected prod-1, got %s", p.ID)
	}
}

func TestStore_GetNotFound(t *testing.T) {
	s := NewStore()
	_, err := s.Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent product")
	}
}
