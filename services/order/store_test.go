package order

import (
	"context"
	"testing"
	"time"

	"github.com/erikeverts/observability-test-app/internal/model"
)

func TestStore_SaveAndGet(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()
	order := &model.Order{
		ID:        "order-1",
		Items:     []model.OrderItem{{ProductID: "prod-1", Quantity: 2}},
		Status:    model.OrderStatusConfirmed,
		Total:     299.98,
		CreatedAt: time.Now(),
	}

	if err := s.Save(ctx, order); err != nil {
		t.Fatalf("unexpected save error: %v", err)
	}
	got, err := s.Get(ctx, "order-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "order-1" {
		t.Errorf("expected order-1, got %s", got.ID)
	}
	if got.Total != 299.98 {
		t.Errorf("expected total 299.98, got %f", got.Total)
	}
}

func TestStore_GetNotFound(t *testing.T) {
	s := NewMemoryStore()
	_, err := s.Get(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent order")
	}
}

func TestStore_List(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()
	s.Save(ctx, &model.Order{ID: "order-1", Status: model.OrderStatusConfirmed, CreatedAt: time.Now()})
	s.Save(ctx, &model.Order{ID: "order-2", Status: model.OrderStatusPending, CreatedAt: time.Now()})

	orders, err := s.List(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orders) != 2 {
		t.Errorf("expected 2 orders, got %d", len(orders))
	}
}
