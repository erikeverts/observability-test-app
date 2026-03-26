package chaos

import (
	"log/slog"
	"math/rand"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ErrorInjector struct {
	routes map[string]float64
}

func NewErrorInjector(routes map[string]float64) *ErrorInjector {
	return &ErrorInjector{routes: routes}
}

func (e *ErrorInjector) Middleware(next http.Handler) http.Handler {
	if len(e.routes) == 0 {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rate, ok := e.routes[r.URL.Path]
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
			http.Error(w, `{"error":"chaos: injected error"}`, http.StatusInternalServerError)
			return
		}
		next.ServeHTTP(w, r)
	})
}
