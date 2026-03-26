package chaos

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/erikeverts/observability-test-app/internal/config"
)

type ChaosConfig struct {
	ErrorRoutes   map[string]float64 `json:"error_routes"`
	LatencyRoutes map[string]string  `json:"latency_routes"`
	CPULoad       CPULoadConfig      `json:"cpu_load"`
	MemLoad       MemLoadConfig      `json:"mem_load"`
	LogVolume     LogVolumeConfig    `json:"log_volume"`
}

type CPULoadConfig struct {
	Enabled bool `json:"enabled"`
	Percent int  `json:"percent"`
}

type MemLoadConfig struct {
	Enabled bool `json:"enabled"`
	MB      int  `json:"mb"`
}

type LogVolumeConfig struct {
	Enabled    bool   `json:"enabled"`
	RatePerSec int    `json:"rate_per_sec"`
	Pattern    string `json:"pattern"`
}

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

func (c *Chaos) GetConfig() ChaosConfig {
	latencyRoutes := c.LatencyInjector.GetRoutes()
	latencyStrs := make(map[string]string, len(latencyRoutes))
	for path, d := range latencyRoutes {
		latencyStrs[path] = d.String()
	}

	cpuEnabled, cpuPercent, memEnabled, memMB := c.LoadSimulator.GetConfig()
	logEnabled, logRate, logPattern := c.LogGenerator.GetConfig()

	return ChaosConfig{
		ErrorRoutes:   c.ErrorInjector.GetRoutes(),
		LatencyRoutes: latencyStrs,
		CPULoad:       CPULoadConfig{Enabled: cpuEnabled, Percent: cpuPercent},
		MemLoad:       MemLoadConfig{Enabled: memEnabled, MB: memMB},
		LogVolume:     LogVolumeConfig{Enabled: logEnabled, RatePerSec: logRate, Pattern: logPattern},
	}
}

func (c *Chaos) ApplyConfig(cfg ChaosConfig) error {
	c.ErrorInjector.SetRoutes(cfg.ErrorRoutes)

	latencyRoutes := make(map[string]time.Duration)
	for path, durStr := range cfg.LatencyRoutes {
		d, err := time.ParseDuration(durStr)
		if err != nil {
			return fmt.Errorf("invalid duration %q for route %s: %w", durStr, path, err)
		}
		latencyRoutes[path] = d
	}
	c.LatencyInjector.SetRoutes(latencyRoutes)

	c.LoadSimulator.Reconfigure(cfg.CPULoad.Enabled, cfg.CPULoad.Percent, cfg.MemLoad.Enabled, cfg.MemLoad.MB)
	c.LogGenerator.Reconfigure(cfg.LogVolume.Enabled, cfg.LogVolume.RatePerSec, cfg.LogVolume.Pattern)

	return nil
}

func (c *Chaos) HandleGetConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c.GetConfig())
}

func (c *Chaos) HandleSetConfig(w http.ResponseWriter, r *http.Request) {
	var cfg ChaosConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	if err := c.ApplyConfig(cfg); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c.GetConfig())
}
