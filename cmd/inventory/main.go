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
	"github.com/erikeverts/observability-test-app/internal/database"
	"github.com/erikeverts/observability-test-app/internal/middleware"
	"github.com/erikeverts/observability-test-app/internal/telemetry"
	"github.com/erikeverts/observability-test-app/services/inventory"
	"github.com/jackc/pgx/v5/pgxpool"
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

	var pool *pgxpool.Pool
	if cfg.DatabaseURL != "" {
		var dbErr error
		pool, dbErr = database.NewPool(ctx, cfg.DatabaseURL)
		if dbErr != nil {
			slog.Error("failed to connect to database", "error", dbErr)
			os.Exit(1)
		}
		defer pool.Close()
		if dbErr := database.RunMigrations(ctx, pool); dbErr != nil {
			slog.Error("failed to run migrations", "error", dbErr)
			os.Exit(1)
		}
	}

	svc, err := inventory.NewService(cfg.InventoryDataDir, pool)
	if err != nil {
		slog.Error("failed to init inventory service", "error", err)
		os.Exit(1)
	}

	c := chaos.New(cfg)
	c.LoadSimulator.Start(ctx)
	c.DiskFiller.Start(ctx)
	go c.LogGenerator.Start(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		if pool != nil {
			if err := pool.Ping(r.Context()); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(w).Encode(map[string]string{"status": "not ready", "error": "database unavailable"})
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	})
	mux.HandleFunc("GET /admin/chaos", c.HandleGetConfig)
	mux.HandleFunc("PUT /admin/chaos", c.HandleSetConfig)
	mux.Handle("/", svc.Mux)

	handler := c.Middleware(middleware.Wrap(mux, cfg.ServiceName))

	addr := ":" + cfg.ServicePort
	slog.Info("starting inventory service", "addr", addr, "data_dir", cfg.InventoryDataDir)
	if err := http.ListenAndServe(addr, handler); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
