package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/example/starterapp/backend/internal/optional/livefeed"
)

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	// /health and the unauthenticated paths below never touch Postgres, so a
	// nil pool is safe here.
	return setupRouter(nil, []byte("test-only-fixed-32-byte-secret-k"), livefeed.NewHub())
}

func TestHealthReturns200OK(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "/health", nil)
	newTestRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	const expected = `{"status":"ok"}`
	if w.Body.String() != expected {
		t.Fatalf("expected body %q, got %q", expected, w.Body.String())
	}
}

func TestReadyReturns503WithoutADatabase(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "/ready", nil)
	newTestRouter().ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}
}

func TestEventsStreamRequiresASession(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/events", nil)
	newTestRouter().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}
