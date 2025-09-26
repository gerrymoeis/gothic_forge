package main

import (
    "fmt"
    "log"

    "github.com/gofiber/fiber/v2"
    "gothicforge/app/routes"
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
					html := fmt.Sprintf("<!doctype html><html><head><meta charset=\"utf-8\"><title>%d</title></head><body><h1>%d</h1><p>%s</p></body></html>", fe.Code, fe.Code, fe.Message)
					return c.SendString(html)
				}
				return err
			}
			c.Type("html", "utf-8")
			c.Status(fiber.StatusInternalServerError)
			html := fmt.Sprintf("<!doctype html><html><head><meta charset=\"utf-8\"><title>%d</title></head><body><h1>%d</h1><p>%s</p></body></html>", fiber.StatusInternalServerError, fiber.StatusInternalServerError, err.Error())
			return c.SendString(html)
		}
		return nil
	})

	routes.Register(app)

	// Catch-all fallback 404 route (placed last). No Next() call, ensures HTML 404.
	app.All("/*", func(c *fiber.Ctx) error {
		c.Type("html", "utf-8")
		c.Status(fiber.StatusNotFound)
		html := "<!doctype html><html><head><meta charset=\"utf-8\"><title>404</title></head><body><h1>404</h1><p>Not Found</p></body></html>"
		return c.SendString(html)
	})

	host := env.Get("FIBER_HOST", "127.0.0.1")
	port := env.Get("FIBER_PORT", "8080")
	addr := fmt.Sprintf("%s:%s", host, port)

	log.Printf("Gothic Forge listening at http://%s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
