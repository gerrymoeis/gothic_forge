package main

import (
  "io"
  "net/http/httptest"
  "strings"
  "testing"

  "github.com/gofiber/fiber/v2"
  "gothicforge/app/routes"
  "gothicforge/app/templates"
  "gothicforge/internal/framework/env"
  "gothicforge/internal/framework/server"
)

// buildTestApp mirrors main.go wiring to validate 404 rendering.
func buildTestApp() *fiber.App {
  _ = env.Load()
  app := server.New()
  routes.Register(app)
  // Use the same catch-all 404 fallback as main.go
  app.All("/*", func(c *fiber.Ctx) error {
    c.Type("html", "utf-8")
    c.Status(fiber.StatusNotFound)
    return templates.NotFound().Render(c.UserContext(), c.Response().BodyWriter())
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
