package tests

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode"

	"github.com/gofiber/fiber/v2"
	"gothicforge/app/routes"
	"gothicforge/internal/server"
)

func setupApp() *fiber.App {
	app := server.New()
	routes.Register(app)
	return app
}

func TestHome(t *testing.T) {
	app := setupApp()
	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("home request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Fatalf("expected text/html content-type, got %q", ct)
	}
}

func TestFaviconRedirect(t *testing.T) {
	app := setupApp()
	req := httptest.NewRequest("GET", "/favicon.ico", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("favicon request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusMovedPermanently {
		t.Fatalf("expected 301, got %d", resp.StatusCode)
	}
	loc := resp.Header.Get("Location")
	if loc != "/static/favicon.svg" {
		t.Fatalf("expected redirect to /static/favicon.svg, got %q", loc)
	}
}

func TestRobotsRedirect(t *testing.T) {
	app := setupApp()
	req := httptest.NewRequest("GET", "/robots.txt", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("robots request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusMovedPermanently {
		t.Fatalf("expected 301, got %d", resp.StatusCode)
	}
	loc := resp.Header.Get("Location")
	if loc != "/static/robots.txt" {
		t.Fatalf("expected redirect to /static/robots.txt, got %q", loc)
	}
}

func TestSitemapRedirect(t *testing.T) {
	app := setupApp()
	req := httptest.NewRequest("GET", "/sitemap.xml", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("sitemap request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusMovedPermanently {
		t.Fatalf("expected 301, got %d", resp.StatusCode)
	}
	loc := resp.Header.Get("Location")
	if loc != "/static/sitemap.xml" {
		t.Fatalf("expected redirect to /static/sitemap.xml, got %q", loc)
	}
}

func extractCookiePair(setCookie string, name string) string {
	parts := strings.Split(setCookie, ";")
	if len(parts) > 0 && strings.HasPrefix(strings.TrimSpace(parts[0]), name+"=") {
		return strings.TrimSpace(parts[0])
	}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, name+"=") {
			if strings.Contains(p, "=") && !strings.ContainsAny(p, " ") && !strings.HasPrefix(p, "Path=") {
				return p
			}
		}
	}
	return ""
}

func TestCounterIncrement_POST_Once(t *testing.T) {
	app := setupApp()
	// Seed CSRF cookie via GET /
	get := httptest.NewRequest("GET", "/", nil)
	got, err := app.Test(get)
	if err != nil {
		t.Fatalf("seed GET failed: %v", err)
	}
	setCookie := got.Header.Get("Set-Cookie")
	csrfPair := extractCookiePair(setCookie, "_gforge_csrf")
	sessPair := extractCookiePair(setCookie, "session")
	if csrfPair == "" {
		t.Fatalf("missing csrf cookie in response")
	}

	req := httptest.NewRequest("POST", "/counter/increment", nil)
	token := strings.SplitN(csrfPair, "=", 2)[1]
	req.Header.Set("X-CSRF-Token", token)
	if sessPair != "" {
		req.Header.Set("Cookie", sessPair+"; "+csrfPair)
	} else {
		req.Header.Set("Cookie", csrfPair)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("/counter/increment POST failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	s := strings.TrimSpace(string(body))
	if s == "" {
		t.Fatalf("empty response")
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			t.Fatalf("expected digits, got: %q", s)
		}
	}
}

func TestCounterIncrement_POST_TwiceWithSession(t *testing.T) {
	app := setupApp()
	// Seed cookies via GET /
	g := httptest.NewRequest("GET", "/", nil)
	gr, err := app.Test(g)
	if err != nil {
		t.Fatalf("seed GET failed: %v", err)
	}
	sc := gr.Header.Get("Set-Cookie")
	csrfPair := extractCookiePair(sc, "_gforge_csrf")
	sessPair := extractCookiePair(sc, "session")
	if csrfPair == "" {
		t.Fatalf("missing csrf cookie in response")
	}

	// First increment
	req1 := httptest.NewRequest("POST", "/counter/increment", nil)
	token := strings.SplitN(csrfPair, "=", 2)[1]
	req1.Header.Set("X-CSRF-Token", token)
	if sessPair != "" {
		req1.Header.Set("Cookie", sessPair+"; "+csrfPair)
	} else {
		req1.Header.Set("Cookie", csrfPair)
	}
	resp1, err := app.Test(req1)
	if err != nil {
		t.Fatalf("first increment failed: %v", err)
	}
	// Capture session cookie
	sessCookie := resp1.Header.Get("Set-Cookie")
	// Second increment with session + csrf
	req2 := httptest.NewRequest("POST", "/counter/increment", nil)
	req2.Header.Set("X-CSRF-Token", token)
	// merge session cookie (may be refreshed) with csrfPair
	if sessCookie != "" {
		semi := strings.Index(sessCookie, ";")
		if semi > 0 {
			sessCookie = sessCookie[:semi]
		}
		req2.Header.Set("Cookie", sessCookie+"; "+csrfPair)
	} else if sessPair != "" {
		req2.Header.Set("Cookie", sessPair+"; "+csrfPair)
	} else {
		req2.Header.Set("Cookie", csrfPair)
	}
	resp2, err := app.Test(req2)
	if err != nil {
		t.Fatalf("second increment failed: %v", err)
	}
	if resp2.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp2.StatusCode)
	}
	body2, _ := io.ReadAll(resp2.Body)
	s2 := strings.TrimSpace(string(body2))
	if s2 == "" {
		t.Fatalf("empty response")
	}
	for _, r := range s2 {
		if !unicode.IsDigit(r) {
			t.Fatalf("expected digits, got: %q", s2)
		}
	}
}
