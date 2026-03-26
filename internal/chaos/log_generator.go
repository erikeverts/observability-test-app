package chaos

import (
	"context"
	"log/slog"
	"math/rand"
	"time"
)

type LogGenerator struct {
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

func (g *LogGenerator) Start(ctx context.Context) {
	if !g.enabled || g.ratePerSec <= 0 {
		return
	}
	slog.Info("chaos: starting log generation", "rate_per_sec", g.ratePerSec, "pattern", g.pattern)

	interval := time.Second / time.Duration(g.ratePerSec)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			g.emitLog()
		}
	}
}

func (g *LogGenerator) emitLog() {
	attrs := []any{
		"source", "chaos_log_generator",
		"pattern", g.pattern,
		"random_id", rand.Intn(10000),
	}

	switch g.pattern {
	case "error":
		slog.Error("chaos: generated error log", attrs...)
	case "warn":
		slog.Warn("chaos: generated warning log", attrs...)
	case "info":
		slog.Info("chaos: generated info log", attrs...)
	default: // "mixed"
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
