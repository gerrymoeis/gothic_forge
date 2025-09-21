package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"gothicforge/app/routes"
	"gothicforge/app/templates"
	"gothicforge/internal/env"
	"gothicforge/internal/server"
)

func main() {
	_ = env.Load()
	app := server.New()

	// Error handler middleware to render HTML for server errors (>=500).
	// Must be registered BEFORE routes so it can catch downstream errors.
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

	routes.Register(app)

	// Catch-all fallback 404 route (placed last). No Next() call, ensures HTML 404.
	app.All("/*", func(c *fiber.Ctx) error {
		c.Type("html", "utf-8")
		c.Status(fiber.StatusNotFound)
		return templates.NotFound().Render(c.UserContext(), c.Response().BodyWriter())
	})

	host := env.Get("FIBER_HOST", "127.0.0.1")
	port := env.Get("FIBER_PORT", "8080")
	addr := fmt.Sprintf("%s:%s", host, port)

	log.Printf("Gothic Forge listening at http://%s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
