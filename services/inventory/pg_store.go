package inventory

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGStore struct {
	pool       *pgxpool.Pool
	dataDir    string
	ledgerFile *os.File
}

func NewPGStore(pool *pgxpool.Pool, dataDir string) (*PGStore, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("creating data dir: %w", err)
	}

	ledgerPath := filepath.Join(dataDir, "ledger.jsonl")
	f, err := os.OpenFile(ledgerPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("opening ledger: %w", err)
	}

	return &PGStore{
		pool:       pool,
		dataDir:    dataDir,
		ledgerFile: f,
	}, nil
}

func (s *PGStore) GetStock(ctx context.Context, productID string) (int, bool) {
	var qty int
	err := s.pool.QueryRow(ctx,
		`SELECT quantity FROM inventory WHERE product_id = $1`, productID).Scan(&qty)
	if err != nil {
		return 0, false
	}
	return qty, true
}

func (s *PGStore) ListStock(ctx context.Context) ([]StockEntry, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT product_id, quantity FROM inventory ORDER BY product_id`)
	if err != nil {
		return nil, fmt.Errorf("querying inventory: %w", err)
	}
	defer rows.Close()

	var entries []StockEntry
	for rows.Next() {
		var e StockEntry
		if err := rows.Scan(&e.ProductID, &e.Quantity); err != nil {
			return nil, fmt.Errorf("scanning inventory: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func (s *PGStore) Reserve(ctx context.Context, productID string, quantity int) (int, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var current int
	err = tx.QueryRow(ctx,
		`SELECT quantity FROM inventory WHERE product_id = $1 FOR UPDATE`, productID).Scan(&current)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("product %s not found in inventory", productID)
		}
		return 0, fmt.Errorf("querying stock: %w", err)
	}

	if current < quantity {
		return current, fmt.Errorf("insufficient stock for %s: have %d, need %d", productID, current, quantity)
	}

	remaining := current - quantity
	_, err = tx.Exec(ctx,
		`UPDATE inventory SET quantity = $1 WHERE product_id = $2`, remaining, productID)
	if err != nil {
		return 0, fmt.Errorf("updating stock: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("committing reservation: %w", err)
	}

	s.appendLedger(productID, "reserve", -quantity, remaining)
	return remaining, nil
}

func (s *PGStore) DiskUsage() (int64, error) {
	var total int64
	entries, err := os.ReadDir(s.dataDir)
	if err != nil {
		return 0, err
	}
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		total += info.Size()
	}
	return total, nil
}

func (s *PGStore) appendLedger(productID, action string, delta, remaining int) {
	entry := LedgerEntry{
		Timestamp: time.Now(),
		ProductID: productID,
		Action:    action,
		Delta:     delta,
		Remaining: remaining,
	}
	data, _ := json.Marshal(entry)
	data = append(data, '\n')
	s.ledgerFile.Write(data)
}
