package telemetry

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

func NewLogger(lp *sdklog.LoggerProvider, serviceName string) *slog.Logger {
	otelHandler := otelslog.NewHandler(serviceName, otelslog.WithLoggerProvider(lp))

	return slog.New(fanoutHandler{
		handlers: []slog.Handler{
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
			otelHandler,
		},
	})
}

// fanoutHandler sends log records to multiple handlers.
type fanoutHandler struct {
	handlers []slog.Handler
}

func (f fanoutHandler) Enabled(_ context.Context, level slog.Level) bool {
	for _, h := range f.handlers {
		if h.Enabled(nil, level) {
			return true
		}
	}
	return false
}

func (f fanoutHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range f.handlers {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r); err != nil {
				return err
			}
		}
	}
	return nil
}

func (f fanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(f.handlers))
	for i, h := range f.handlers {
		handlers[i] = h.WithAttrs(attrs)
	}
	return fanoutHandler{handlers: handlers}
}

func (f fanoutHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(f.handlers))
	for i, h := range f.handlers {
		handlers[i] = h.WithGroup(name)
	}
	return fanoutHandler{handlers: handlers}
}
