package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/erikeverts/observability-test-app/internal/chaos"
	"github.com/erikeverts/observability-test-app/internal/config"
	"github.com/erikeverts/observability-test-app/internal/middleware"
	"github.com/erikeverts/observability-test-app/internal/telemetry"
	"github.com/erikeverts/observability-test-app/services/gateway"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	cfg := config.Load()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	tel, cleanup, err := telemetry.InitTelemetry(ctx, cfg)
	if err != nil {
		slog.Error("failed to init telemetry", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	profilingCleanup, err := telemetry.InitProfiling(cfg)
	if err != nil {
		slog.Error("failed to init profiling", "error", err)
		os.Exit(1)
	}
	defer profilingCleanup()

	logger := telemetry.NewLogger(tel.LoggerProvider, cfg.ServiceName)
	slog.SetDefault(logger)

	httpClient := &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	svc := gateway.NewService(cfg.ProductServiceURL, cfg.OrderServiceURL, cfg.InventoryServiceURL, httpClient)

	metrics, err := telemetry.NewAppMetrics()
	if err != nil {
		slog.Error("failed to init app metrics", "error", err)
		os.Exit(1)
	}

	c := chaos.New(cfg, metrics.ChaosErrorsTotal, metrics.ChaosLatencyTotal)
	c.LoadSimulator.Start(ctx)
	go c.LogGenerator.Start(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	})
	mux.HandleFunc("GET /admin/chaos", c.HandleGetConfig)
	mux.HandleFunc("PUT /admin/chaos", c.HandleSetConfig)
	mux.Handle("/", svc.Mux)

	handler := c.Middleware(middleware.Wrap(mux, cfg.ServiceName))

	addr := ":" + cfg.ServicePort
	slog.Info("starting gateway", "addr", addr, "product_service", cfg.ProductServiceURL, "order_service", cfg.OrderServiceURL, "inventory_service", cfg.InventoryServiceURL)
	if err := http.ListenAndServe(addr, handler); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
