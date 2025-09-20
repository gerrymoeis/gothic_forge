package main

import (
  "io"
  "net/http/httptest"
  "strings"
  "testing"

  "github.com/gofiber/fiber/v2"
  "gothicforge/app/templates"
  "gothicforge/internal/framework/env"
  "gothicforge/internal/framework/server"
)

func buildErrorTestApp() *fiber.App {
  _ = env.Load()
  app := server.New()

  // Error handler middleware mirroring main.go (register BEFORE routes)
  app.Use(func(c *fiber.Ctx) error {
    if err := c.Next(); err != nil {
      if fe, ok := err.(*fiber.Error); ok {
        if fe.Code >= 500 {
          c.Type("html", "utf-8")
          c.Status(fe.Code)
          return templates.ErrorPage(fe.Code, fe.Message).Render(c.UserContext(), c.Response().BodyWriter())
        }
        return err
      }
      c.Type("html", "utf-8")
      c.Status(fiber.StatusInternalServerError)
      return templates.ErrorPage(fiber.StatusInternalServerError, err.Error()).Render(c.UserContext(), c.Response().BodyWriter())
    }
    return nil
  })

  // Route that triggers a 500 error
  app.Get("/boom", func(c *fiber.Ctx) error {
    return fiber.NewError(fiber.StatusInternalServerError, "boom")
  })

  // 404 fallback for completeness
  app.All("/*", func(c *fiber.Ctx) error {
    c.Type("html", "utf-8")
    c.Status(fiber.StatusNotFound)
    return templates.NotFound().Render(c.UserContext(), c.Response().BodyWriter())
  })

  return app
}

func TestErrorPage_HTML500(t *testing.T) {
  app := buildErrorTestApp()
  req := httptest.NewRequest("GET", "/boom", nil)
  resp, err := app.Test(req)
  if err != nil { t.Fatalf("request failed: %v", err) }
  if resp.StatusCode != fiber.StatusInternalServerError {
    t.Fatalf("expected 500, got %d", resp.StatusCode)
  }
  ct := resp.Header.Get("Content-Type")
  if !strings.Contains(ct, "text/html") {
    t.Fatalf("expected text/html, got %q", ct)
  }
  body, _ := io.ReadAll(resp.Body)
  s := string(body)
  if !strings.Contains(s, "500") || !strings.Contains(s, "boom") {
    t.Fatalf("expected error page content, got: %s", s)
  }
}
