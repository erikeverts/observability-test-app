package order

import (
	"context"

	"github.com/erikeverts/observability-test-app/internal/model"
)

type OrderStore interface {
	Save(ctx context.Context, order *model.Order) error
	Get(ctx context.Context, id string) (*model.Order, error)
	List(ctx context.Context) ([]model.Order, error)
}
