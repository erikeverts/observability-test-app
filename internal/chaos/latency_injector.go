package chaos

import (
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type LatencyInjector struct {
	routes map[string]time.Duration
}

func NewLatencyInjector(routes map[string]time.Duration) *LatencyInjector {
	return &LatencyInjector{routes: routes}
}

func (l *LatencyInjector) Middleware(next http.Handler) http.Handler {
	if len(l.routes) == 0 {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		delay, ok := l.routes[r.URL.Path]
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		span := trace.SpanFromContext(r.Context())
		span.SetAttributes(
			attribute.Bool("chaos.latency_injected", true),
			attribute.String("chaos.latency_delay", delay.String()),
		)
		slog.WarnContext(r.Context(), "chaos: injecting latency",
			"path", r.URL.Path,
			"delay", delay.String(),
		)
		time.Sleep(delay)
		next.ServeHTTP(w, r)
	})
}
