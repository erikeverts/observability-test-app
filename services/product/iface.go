package product

import (
	"context"

	"github.com/erikeverts/observability-test-app/internal/model"
)

type ProductStore interface {
	List(ctx context.Context) ([]model.Product, error)
	Get(ctx context.Context, id string) (*model.Product, error)
}
