package chaos

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type LatencyInjector struct {
	mu      sync.RWMutex
	routes  map[string]time.Duration
	counter metric.Int64Counter
}

func NewLatencyInjector(routes map[string]time.Duration, counter metric.Int64Counter) *LatencyInjector {
	return &LatencyInjector{routes: routes, counter: counter}
}

func (l *LatencyInjector) SetRoutes(routes map[string]time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.routes = routes
}

func (l *LatencyInjector) GetRoutes() map[string]time.Duration {
	l.mu.RLock()
	defer l.mu.RUnlock()
	cp := make(map[string]time.Duration, len(l.routes))
	for k, v := range l.routes {
		cp[k] = v
	}
	return cp
}

func (l *LatencyInjector) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l.mu.RLock()
		delay, ok := l.routes[r.URL.Path]
		l.mu.RUnlock()

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
		l.counter.Add(r.Context(), 1)
		time.Sleep(delay)
		next.ServeHTTP(w, r)
	})
}
