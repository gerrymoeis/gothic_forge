package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"gothicforge/app/routes"
	"gothicforge/app/templates"
	"gothicforge/internal/framework/env"
	"gothicforge/internal/framework/server"
)

func main() {
	_ = env.Load()
	app := server.New()
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
