package chaos

import (
	"net/http"

	"github.com/erikeverts/observability-test-app/internal/config"
)

type Chaos struct {
	ErrorInjector   *ErrorInjector
	LatencyInjector *LatencyInjector
	LoadSimulator   *LoadSimulator
	LogGenerator    *LogGenerator
}

func New(cfg *config.Config) *Chaos {
	return &Chaos{
		ErrorInjector:   NewErrorInjector(cfg.ChaosErrorRoutes),
		LatencyInjector: NewLatencyInjector(cfg.ChaosLatencyRoutes),
		LoadSimulator:   NewLoadSimulator(cfg.ChaosCPULoadEnabled, cfg.ChaosCPULoadPercent, cfg.ChaosMemLoadEnabled, cfg.ChaosMemLoadMB),
		LogGenerator:    NewLogGenerator(cfg.ChaosLogVolumeEnabled, cfg.ChaosLogRatePerSec, cfg.ChaosLogPattern),
	}
}

func (c *Chaos) Middleware(next http.Handler) http.Handler {
	h := next
	h = c.LatencyInjector.Middleware(h)
	h = c.ErrorInjector.Middleware(h)
	return h
}
