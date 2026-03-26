package inventory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type StockEntry struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type LedgerEntry struct {
	Timestamp time.Time `json:"timestamp"`
	ProductID string    `json:"product_id"`
	Action    string    `json:"action"`
	Delta     int       `json:"delta"`
	Remaining int       `json:"remaining"`
}

type Store struct {
	mu        sync.RWMutex
	stock     map[string]int
	dataDir   string
	ledgerFile *os.File
}

func NewStore(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("creating data dir: %w", err)
	}

	ledgerPath := filepath.Join(dataDir, "ledger.jsonl")
	f, err := os.OpenFile(ledgerPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("opening ledger: %w", err)
	}

	s := &Store{
		stock:      make(map[string]int),
		dataDir:    dataDir,
		ledgerFile: f,
	}
	s.seed()
	return s, nil
}

func (s *Store) seed() {
	defaults := map[string]int{
		"prod-1": 50,
		"prod-2": 120,
		"prod-3": 200,
		"prod-4": 35,
		"prod-5": 80,
	}
	for id, qty := range defaults {
		s.stock[id] = qty
		s.appendLedger(id, "seed", qty, qty)
	}
}

func (s *Store) GetStock(productID string) (int, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	qty, ok := s.stock[productID]
	return qty, ok
}

func (s *Store) ListStock() []StockEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entries := make([]StockEntry, 0, len(s.stock))
	for id, qty := range s.stock {
		entries = append(entries, StockEntry{ProductID: id, Quantity: qty})
	}
	return entries
}

func (s *Store) Reserve(productID string, quantity int) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	current, ok := s.stock[productID]
	if !ok {
		return 0, fmt.Errorf("product %s not found in inventory", productID)
	}
	if current < quantity {
		return current, fmt.Errorf("insufficient stock for %s: have %d, need %d", productID, current, quantity)
	}

	s.stock[productID] = current - quantity
	s.appendLedger(productID, "reserve", -quantity, s.stock[productID])
	return s.stock[productID], nil
}

func (s *Store) DiskUsage() (int64, error) {
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

func (s *Store) appendLedger(productID, action string, delta, remaining int) {
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
