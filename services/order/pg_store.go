package order

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

func (s *PGStore) Save(ctx context.Context, order *model.Order) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO orders (id, status, total, created_at) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (id) DO UPDATE SET status = $2, total = $3`,
		order.ID, string(order.Status), order.Total, order.CreatedAt)
	if err != nil {
		return fmt.Errorf("inserting order: %w", err)
	}

	for _, item := range order.Items {
		_, err = tx.Exec(ctx,
			`INSERT INTO order_items (order_id, product_id, quantity) VALUES ($1, $2, $3)`,
			order.ID, item.ProductID, item.Quantity)
		if err != nil {
			return fmt.Errorf("inserting order item: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (s *PGStore) Get(ctx context.Context, id string) (*model.Order, error) {
	var o model.Order
	err := s.pool.QueryRow(ctx,
		`SELECT id, status, total, created_at FROM orders WHERE id = $1`, id).
		Scan(&o.ID, &o.Status, &o.Total, &o.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("order %s not found", id)
		}
		return nil, fmt.Errorf("querying order: %w", err)
	}

	rows, err := s.pool.Query(ctx,
		`SELECT product_id, quantity FROM order_items WHERE order_id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("querying order items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item model.OrderItem
		if err := rows.Scan(&item.ProductID, &item.Quantity); err != nil {
			return nil, fmt.Errorf("scanning order item: %w", err)
		}
		o.Items = append(o.Items, item)
	}
	return &o, nil
}

func (s *PGStore) List(ctx context.Context) ([]model.Order, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, status, total, created_at FROM orders ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("querying orders: %w", err)
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.Status, &o.Total, &o.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning order: %w", err)
		}
		orders = append(orders, o)
	}

	for i := range orders {
		itemRows, err := s.pool.Query(ctx,
			`SELECT product_id, quantity FROM order_items WHERE order_id = $1`, orders[i].ID)
		if err != nil {
			return nil, fmt.Errorf("querying items for order %s: %w", orders[i].ID, err)
		}
		for itemRows.Next() {
			var item model.OrderItem
			if err := itemRows.Scan(&item.ProductID, &item.Quantity); err != nil {
				itemRows.Close()
				return nil, fmt.Errorf("scanning order item: %w", err)
			}
			orders[i].Items = append(orders[i].Items, item)
		}
		itemRows.Close()
	}
	return orders, nil
}
