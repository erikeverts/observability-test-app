package database

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schemaSQL string

func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing database URL: %w", err)
	}

	cfg.ConnConfig.Tracer = otelpgx.NewTracer(
		otelpgx.WithIncludeQueryParameters(),
	)

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	slog.Info("database connection established",
		"host", cfg.ConnConfig.Host,
		"database", cfg.ConnConfig.Database,
	)
	return pool, nil
}

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, schemaSQL)
	if err != nil {
		return fmt.Errorf("running schema migrations: %w", err)
	}
	slog.Info("database migrations completed")
	return nil
}
