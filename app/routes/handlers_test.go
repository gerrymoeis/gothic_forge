package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gothicforge3/internal/server"
)

func TestHealthz(t *testing.T) {
	r := server.New()
	Register(r)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}

func TestFaviconRedirect(t *testing.T) {
	r := server.New()
	Register(r)
	req := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("want 301, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc != "/static/favicon.svg" {
		t.Fatalf("unexpected redirect: %q", loc)
	}
}

func TestCounterSync(t *testing.T) {
	r := server.New()
	Register(r)
	form := "count=7"
	req := httptest.NewRequest(http.MethodPost, "/counter/sync", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if rec.Body.String() != "7" {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}
