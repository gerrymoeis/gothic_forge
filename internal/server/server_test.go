package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestDetectStaticDir_RespectsBaseDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("GFORGE_BASEDIR", tmp)
	got := detectStaticDir()
	want := filepath.Join(tmp, "app", "static")
	if got != want {
		t.Fatalf("detectStaticDir: want %q, got %q", want, got)
	}
}

func TestCSRFMiddleware_AllowsSafeAndBlocksCrossOrigin(t *testing.T) {
	h := CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	// GET allowed
	req := httptest.NewRequest(http.MethodGet, "http://example.test/x", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("GET: want 200, got %d", rec.Code)
	}
	// POST without Origin/Referer -> 403
	req2 := httptest.NewRequest(http.MethodPost, "http://example.test/x", nil)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusForbidden {
		t.Fatalf("POST no origin: want 403, got %d", rec2.Code)
	}
	// POST with mismatched Origin -> 403
	req3 := httptest.NewRequest(http.MethodPost, "http://example.test/x", nil)
	req3.Header.Set("Origin", "http://other")
	rec3 := httptest.NewRecorder()
	h.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusForbidden {
		t.Fatalf("POST bad origin: want 403, got %d", rec3.Code)
	}
	// POST with matching Origin -> 200
	req4 := httptest.NewRequest(http.MethodPost, "http://example.test/x", nil)
	req4.Header.Set("Origin", "http://example.test")
	rec4 := httptest.NewRecorder()
	h.ServeHTTP(rec4, req4)
	if rec4.Code != 200 {
		t.Fatalf("POST good origin: want 200, got %d", rec4.Code)
	}
}

func TestCORS_DefaultAndExplicitOrigins(t *testing.T) {
	// Default: *
	mw := configureCORS()
	b := &bytes.Buffer{}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	req := httptest.NewRequest(http.MethodGet, "http://ex/", nil)
	req.Header.Set("Origin", "http://foo")
	rec := httptest.NewRecorder()
	mw(next).ServeHTTP(rec, req)
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("default CORS: want '*', got %q", got)
	}
	_ = b
	// Explicit origins
	t.Setenv("CORS_ORIGINS", "https://a.com, https://b.com")
	mw2 := configureCORS()
	req2 := httptest.NewRequest(http.MethodGet, "http://ex/", nil)
	req2.Header.Set("Origin", "https://a.com")
	rec2 := httptest.NewRecorder()
	mw2(next).ServeHTTP(rec2, req2)
	if got := rec2.Header().Get("Access-Control-Allow-Origin"); got != "https://a.com" {
		t.Fatalf("explicit CORS: want 'https://a.com', got %q", got)
	}
	if cred := rec2.Header().Get("Access-Control-Allow-Credentials"); cred != "true" {
		t.Fatalf("explicit CORS: want credentials true, got %q", cred)
	}
}
