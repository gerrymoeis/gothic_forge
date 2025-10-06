package tests

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"gothicforge3/app/routes"
	"gothicforge3/internal/auth"
	"gothicforge3/internal/server"
)

func Test_API_Me_Unauthorized(t *testing.T) {
	_ = os.Setenv("LOG_FORMAT", "off")
	r := server.New()
	routes.Register(r)
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
}

func Test_API_Me_Authorized(t *testing.T) {
	_ = os.Setenv("LOG_FORMAT", "off")
	_ = os.Setenv("JWT_SECRET", "testsecret")
	auth.Init()
	tok, _, err := auth.Issue(1*time.Hour, map[string]any{"sub": 1, "login": "tester"})
	if err != nil { t.Fatalf("issue token: %v", err) }

	r := server.New()
	routes.Register(r)
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.AddCookie(&http.Cookie{Name: "gf_jwt", Value: tok, Path: "/"})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "tester") {
		t.Fatalf("expected body to contain 'tester', got %q", rec.Body.String())
	}
}
