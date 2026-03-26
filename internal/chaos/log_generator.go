package chaos

import (
	"context"
	"log/slog"
	"math/rand"
	"sync"
	"time"
)

type LogGenerator struct {
	mu        sync.Mutex
	parentCtx context.Context
	cancel    context.CancelFunc

	enabled    bool
	ratePerSec int
	pattern    string
}

func NewLogGenerator(enabled bool, ratePerSec int, pattern string) *LogGenerator {
	return &LogGenerator{
		enabled:    enabled,
		ratePerSec: ratePerSec,
		pattern:    pattern,
	}
}

func (g *LogGenerator) Start(parentCtx context.Context) {
	g.mu.Lock()
	g.parentCtx = parentCtx
	g.mu.Unlock()
	g.startLocked()
}

func (g *LogGenerator) Reconfigure(enabled bool, ratePerSec int, pattern string) {
	g.mu.Lock()
	if g.cancel != nil {
		g.cancel()
		g.cancel = nil
	}
	g.enabled = enabled
	g.ratePerSec = ratePerSec
	g.pattern = pattern
	g.mu.Unlock()
	g.startLocked()
}

func (g *LogGenerator) GetConfig() (enabled bool, ratePerSec int, pattern string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.enabled, g.ratePerSec, g.pattern
}

func (g *LogGenerator) startLocked() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.parentCtx == nil || !g.enabled || g.ratePerSec <= 0 {
		return
	}

	ctx, cancel := context.WithCancel(g.parentCtx)
	g.cancel = cancel
	rate := g.ratePerSec
	pattern := g.pattern

	slog.Info("chaos: starting log generation", "rate_per_sec", rate, "pattern", pattern)

	go func() {
		interval := time.Second / time.Duration(rate)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				emitLog(pattern)
			}
		}
	}()
}

func emitLog(pattern string) {
	attrs := []any{
		"source", "chaos_log_generator",
		"pattern", pattern,
		"random_id", rand.Intn(10000),
	}
	switch pattern {
	case "error":
		slog.Error("chaos: generated error log", attrs...)
	case "warn":
		slog.Warn("chaos: generated warning log", attrs...)
	case "info":
		slog.Info("chaos: generated info log", attrs...)
	default:
		switch rand.Intn(3) {
		case 0:
			slog.Info("chaos: generated info log", attrs...)
		case 1:
			slog.Warn("chaos: generated warning log", attrs...)
		case 2:
			slog.Error("chaos: generated error log", attrs...)
		}
	}
}
