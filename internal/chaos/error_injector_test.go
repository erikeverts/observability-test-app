package chaos

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel/metric/noop"
)

func TestErrorInjector_NoRoutes(t *testing.T) {
	injector := NewErrorInjector(nil, noop.Int64Counter{})
	handler := injector.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestErrorInjector_AlwaysError(t *testing.T) {
	injector := NewErrorInjector(map[string]float64{"/fail": 1.0}, noop.Int64Counter{})
	handler := injector.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/fail", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestErrorInjector_NeverError(t *testing.T) {
	injector := NewErrorInjector(map[string]float64{"/safe": 0.0}, noop.Int64Counter{})
	handler := injector.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 100; i++ {
		req := httptest.NewRequest(http.MethodGet, "/safe", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 on iteration %d, got %d", i, rec.Code)
		}
	}
}

func TestErrorInjector_UnmatchedRoute(t *testing.T) {
	injector := NewErrorInjector(map[string]float64{"/fail": 1.0}, noop.Int64Counter{})
	handler := injector.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/other", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
