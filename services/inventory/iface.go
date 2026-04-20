package inventory

import "context"

type InventoryStore interface {
	GetStock(ctx context.Context, productID string) (int, bool)
	ListStock(ctx context.Context) ([]StockEntry, error)
	Reserve(ctx context.Context, productID string, quantity int) (int, error)
	Restock(ctx context.Context, productID string, quantity int) (int, error)
	DiskUsage() (int64, error)
}
