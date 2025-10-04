package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gothicforge3/internal/server"
)

func TestSEOHandlers_Defaults(t *testing.T) {
	// Point static dir resolution to a fresh temp base, without changing CWD
	tmp := t.TempDir()
	t.Setenv("GFORGE_BASEDIR", tmp)
	t.Setenv("SITE_BASE_URL", "http://test.local")

	r := server.New()
	Register(r)

	// robots.txt
	req := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("robots status: want 200, got %d", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if ct == "" || ct[:10] != "text/plain" {
		t.Fatalf("robots content-type: want text/plain, got %q", ct)
	}
	if got := rec.Body.String(); got == "" || (got != "" && !contains(got, "http://test.local/sitemap.xml")) {
		t.Fatalf("robots content: want Sitemap with absolute URL, got %q", got)
	}

	// sitemap.xml
	req2 := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("sitemap status: want 200, got %d", rec2.Code)
	}
	ct2 := rec2.Header().Get("Content-Type")
	if ct2 == "" || ct2[:15] != "application/xml" {
		t.Fatalf("sitemap content-type: want application/xml, got %q", ct2)
	}
	if got := rec2.Body.String(); !contains(got, "<urlset") || !contains(got, "<loc>http://test.local/") {
		t.Fatalf("sitemap content: missing expected elements, got %q", got)
	}
}

func contains(s, sub string) bool { return len(s) >= len(sub) && (s == sub || (len(s) > len(sub) && (contains(s[1:], sub) || s[:len(sub)] == sub))) }
