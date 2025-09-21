package routes

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"gothicforge/internal/server"
)

func setupApp() *fiber.App {
	app := server.New()
	Register(app)
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

func TestHealthz(t *testing.T) {
	app := setupApp()
	req := httptest.NewRequest("GET", "/healthz", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("healthz request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "\"ok\":true") && !strings.Contains(string(body), "\"ok\": true") {
		t.Fatalf("expected JSON with ok=true, got %s", string(body))
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
