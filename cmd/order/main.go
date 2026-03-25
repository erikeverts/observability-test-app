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
	"github.com/erikeverts/observability-test-app/services/order"
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

	logger := telemetry.NewLogger(tel.LoggerProvider, cfg.ServiceName)
	slog.SetDefault(logger)

	httpClient := &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	svc := order.NewService(cfg.ProductServiceURL, httpClient)

	c := chaos.New(cfg)
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
	mux.Handle("/", svc.Mux)

	handler := c.Middleware(middleware.Wrap(mux, cfg.ServiceName))

	addr := ":" + cfg.ServicePort
	slog.Info("starting order service", "addr", addr, "product_service", cfg.ProductServiceURL)
	if err := http.ListenAndServe(addr, handler); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
