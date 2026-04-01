package product

import (
	"context"
	"fmt"

	"github.com/erikeverts/observability-test-app/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGStore struct {
	pool *pgxpool.Pool
}

func NewPGStore(pool *pgxpool.Pool) *PGStore {
	return &PGStore{pool: pool}
}

func (s *PGStore) List(ctx context.Context) ([]model.Product, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, name, description, price, stock FROM products ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("querying products: %w", err)
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock); err != nil {
			return nil, fmt.Errorf("scanning product: %w", err)
		}
		products = append(products, p)
	}
	return products, nil
}

func (s *PGStore) Get(ctx context.Context, id string) (*model.Product, error) {
	var p model.Product
	err := s.pool.QueryRow(ctx,
		`SELECT id, name, description, price, stock FROM products WHERE id = $1`, id).
		Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("product %s not found", id)
		}
		return nil, fmt.Errorf("querying product: %w", err)
	}
	return &p, nil
}
