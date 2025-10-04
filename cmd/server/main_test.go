package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gothicforge3/app/routes"
	"gothicforge3/internal/server"
)

// Smoke test: ensure router initialization and route registration succeeds
// and a basic health endpoint responds with 200.
func TestServerInitAndHealth(t *testing.T) {
	r := server.New()
	routes.Register(r)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("/healthz: want 200, got %d", rec.Code)
	}
}
