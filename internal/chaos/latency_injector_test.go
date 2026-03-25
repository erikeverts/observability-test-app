package chaos

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLatencyInjector_NoRoutes(t *testing.T) {
	injector := NewLatencyInjector(nil)
	handler := injector.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	start := time.Now()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if time.Since(start) > 50*time.Millisecond {
		t.Error("unexpected delay with no routes configured")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestLatencyInjector_AddsDelay(t *testing.T) {
	injector := NewLatencyInjector(map[string]time.Duration{"/slow": 100 * time.Millisecond})
	handler := injector.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	start := time.Now()
	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	elapsed := time.Since(start)
	if elapsed < 90*time.Millisecond {
		t.Errorf("expected at least 90ms delay, got %v", elapsed)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
