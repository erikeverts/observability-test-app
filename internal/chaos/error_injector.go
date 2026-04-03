package chaos

import (
	"log/slog"
	"math/rand"
	"net/http"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type ErrorInjector struct {
	mu      sync.RWMutex
	routes  map[string]float64
	counter metric.Int64Counter
}

func NewErrorInjector(routes map[string]float64, counter metric.Int64Counter) *ErrorInjector {
	return &ErrorInjector{routes: routes, counter: counter}
}

func (e *ErrorInjector) SetRoutes(routes map[string]float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.routes = routes
}

func (e *ErrorInjector) GetRoutes() map[string]float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	cp := make(map[string]float64, len(e.routes))
	for k, v := range e.routes {
		cp[k] = v
	}
	return cp
}

func (e *ErrorInjector) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e.mu.RLock()
		rate, ok := e.routes[r.URL.Path]
		e.mu.RUnlock()

		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		if rand.Float64() < rate {
			span := trace.SpanFromContext(r.Context())
			span.SetAttributes(attribute.Bool("chaos.error_injected", true))
			slog.WarnContext(r.Context(), "chaos: injecting error",
				"path", r.URL.Path,
				"rate", rate,
			)
			e.counter.Add(r.Context(), 1)
			http.Error(w, `{"error":"chaos: injected error"}`, http.StatusInternalServerError)
			return
		}
		next.ServeHTTP(w, r)
	})
}
