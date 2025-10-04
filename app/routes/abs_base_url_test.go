package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAbsBaseURL_EnvOverride(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/", nil)
	// Set via internal env package through real env var
	t.Setenv("SITE_BASE_URL", "https://example.com")
	u := absBaseURL(req)
	if u != "https://example.com" {
		t.Fatalf("want env override, got %q", u)
	}
}

func TestAbsBaseURL_HeaderFallback(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.local/", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	u := absBaseURL(req)
	if u != "https://example.local" {
		t.Fatalf("want https header fallback, got %q", u)
	}
}
