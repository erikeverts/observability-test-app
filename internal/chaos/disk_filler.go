package chaos

import (
	"context"
	"crypto/rand"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type DiskFiller struct {
	mu        sync.Mutex
	parentCtx context.Context
	cancel    context.CancelFunc

	enabled bool
	path    string
	rateMB  int
}

func NewDiskFiller(enabled bool, path string, rateMB int) *DiskFiller {
	return &DiskFiller{
		enabled: enabled,
		path:    path,
		rateMB:  rateMB,
	}
}

func (d *DiskFiller) Start(parentCtx context.Context) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.parentCtx = parentCtx
	d.startLocked()
}

func (d *DiskFiller) Reconfigure(enabled bool, path string, rateMB int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.cancel != nil {
		d.cancel()
		d.cancel = nil
	}
	d.enabled = enabled
	d.path = path
	d.rateMB = rateMB
	d.startLocked()
}

func (d *DiskFiller) GetConfig() (enabled bool, path string, rateMB int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.enabled, d.path, d.rateMB
}

func (d *DiskFiller) startLocked() {
	if d.parentCtx == nil || !d.enabled || d.rateMB <= 0 || d.path == "" {
		return
	}

	ctx, cancel := context.WithCancel(d.parentCtx)
	d.cancel = cancel
	path := d.path
	rateMB := d.rateMB

	slog.Info("chaos: starting disk filler", "path", path, "rate_mb_per_sec", rateMB)

	go func() {
		fillDir := filepath.Join(path, "chaos-fill")
		os.MkdirAll(fillDir, 0755)

		chunk := make([]byte, 1024*1024) // 1MB buffer
		interval := time.Second / time.Duration(rateMB)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		seq := 0

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				rand.Read(chunk)
				filePath := filepath.Join(fillDir, "fill.dat")
				f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					slog.Error("chaos: disk filler write failed", "error", err)
					return
				}
				_, err = f.Write(chunk)
				f.Close()
				if err != nil {
					slog.Warn("chaos: disk filler write error (disk may be full)", "error", err)
					return
				}
				seq++
				if seq%10 == 0 {
					slog.Info("chaos: disk filler progress", "written_mb", seq)
				}
			}
		}
	}()
}
