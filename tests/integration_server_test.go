package tests

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gothicforge3/app/routes"
	"gothicforge3/internal/server"
)

func Test_Healthz_OK(t *testing.T) {
	_ = os.Setenv("LOG_FORMAT", "off")
	r := server.New()
	routes.Register(r)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if strings.TrimSpace(rec.Body.String()) != "ok" {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}

func Test_Favicon_Redirect(t *testing.T) {
	_ = os.Setenv("LOG_FORMAT", "off")
	r := server.New()
	routes.Register(r)
	req := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("want 301, got %d", rec.Code)
	}
	if rec.Header().Get("Location") != "/static/favicon.svg" {
		t.Fatalf("unexpected Location: %q", rec.Header().Get("Location"))
	}
}

func Test_Robots_ServesOrGenerates(t *testing.T) {
	_ = os.Setenv("LOG_FORMAT", "off")
	r := server.New()
	routes.Register(r)
	req := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("robots: want 200, got %d", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/plain") {
		t.Fatalf("robots: unexpected content-type: %q", ct)
	}
}

func Test_Sitemap_ServesOrGenerates(t *testing.T) {
    _ = os.Setenv("LOG_FORMAT", "off")
    r := server.New()
    routes.Register(r)
    req := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
    rec := httptest.NewRecorder()
    r.ServeHTTP(rec, req)
    if rec.Code != 200 {
        t.Fatalf("sitemap: want 200, got %d", rec.Code)
    }
    ct := rec.Header().Get("Content-Type")
    if !strings.Contains(ct, "application/xml") {
        t.Fatalf("sitemap: unexpected content-type: %q", ct)
    }
}

func Test_Static_Exists_Defaults(t *testing.T) {
    _ = os.Setenv("LOG_FORMAT", "off")
    _ = os.Setenv("GFORGE_BASEDIR", "")
    // ensure default static is mounted
    r := server.New()
    req := httptest.NewRequest(http.MethodGet, "/static/", nil)
    rec := httptest.NewRecorder()
    r.ServeHTTP(rec, req)
    // directory listing is not enabled; but route should exist and not 404
    if rec.Code == http.StatusNotFound {
        t.Fatalf("static mount seems missing; got 404")
    }
}

func Test_Tailwind_Input_Written_By_Install(t *testing.T) {
    p := filepath.Join("..", "app", "styles", "tailwind.input.css")
    if _, err := os.Stat(p); os.IsNotExist(err) {
        // don't fail hard; guide with message
        t.Skipf("tailwind.input.css not found at %s (run 'gforge install')", p)
    }
}
