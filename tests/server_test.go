package tests

import (
    "io"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/gofiber/fiber/v2"
    "gothicforge/app/routes"
    "gothicforge/internal/env"
    "gothicforge/internal/server"
)

func buildTestApp() *fiber.App {
	_ = env.Load()
	app := server.New()

	// Error handler middleware (minimal HTML), mirrors cmd/server/main.go
	app.Use(func(c *fiber.Ctx) error {
		if err := c.Next(); err != nil {
			if fe, ok := err.(*fiber.Error); ok && fe.Code >= 500 {
				c.Type("html", "utf-8")
				c.Status(fe.Code)
				html := "<!doctype html><html><head><meta charset=\"utf-8\"><title>" +
					strings.TrimSpace(strings.Split(c.Path(), "?")[0]) + "</title></head><body><h1>" +
					strings.TrimPrefix(strings.TrimSpace(strings.Split(c.Path(), "?")[0]), "/") + "</h1><p>" + fe.Message + "</p></body></html>"
				return c.SendString(html)
			}
			return err
		}
		return nil
	})

	routes.Register(app)

	// Route that triggers a 500 error (must be registered BEFORE catch-all)
	app.Get("/boom", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusInternalServerError, "boom")
	})

	// 404 fallback
	app.All("/*", func(c *fiber.Ctx) error {
		c.Type("html", "utf-8")
		c.Status(fiber.StatusNotFound)
		html := "<!doctype html><html><head><meta charset=\"utf-8\"><title>404</title></head><body><h1>404</h1><p>Not Found</p></body></html>"
		return c.SendString(html)
	})

	return app
}

func TestNotFoundHTML(t *testing.T) {
	app := buildTestApp()
	req := httptest.NewRequest("GET", "/__does_not_exist__", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Fatalf("expected text/html, got %q", ct)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "404") {
		t.Fatalf("expected body to contain 404, got: %s", string(body))
	}
}

func TestErrorPage_HTML500(t *testing.T) {
	app := buildTestApp()

	req := httptest.NewRequest("GET", "/boom", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Fatalf("expected text/html, got %q", ct)
	}
	body, _ := io.ReadAll(resp.Body)
	s := string(body)
	if !strings.Contains(s, "boom") {
		t.Fatalf("expected error page content to include 'boom', got: %s", s)
	}
}
