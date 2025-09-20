package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"gothicforge/app/routes"
	"gothicforge/internal/framework/env"
	"gothicforge/internal/framework/server"
)

func main() {
	_ = env.Load()
	app := server.New()
	routes.Register(app)

	// Fallback 404 handler (placed after routes). Calls next first, then handles 404.
	app.Use(func(c *fiber.Ctx) error {
		if err := c.Next(); err != nil {
			return err
		}
		if c.Response().StatusCode() != fiber.StatusNotFound {
			return nil
		}
		c.Type("html", "utf-8")
		return c.Status(404).SendString(`<!DOCTYPE html>
            <html lang="en" data-theme="dark">
              <head>
                <meta charset="UTF-8"/>
                <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
                <title>Not Found · Gothic Forge</title>
                <link rel="icon" href="/static/favicon.svg" type="image/svg+xml"/>
                <link rel="stylesheet" href="/static/styles.css"/>
                <link rel="preconnect" href="https://cdn.jsdelivr.net"/>
                <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/daisyui@4.12.10/dist/full.css"/>
              </head>
              <body class="min-h-screen bg-base-100 text-base-content grid place-items-center">
                <div class="text-center p-8">
                  <h1 class="text-6xl font-extrabold bg-gradient-to-r from-indigo-500 to-pink-500 bg-clip-text text-transparent">404</h1>
                  <p class="mt-4 opacity-80">Halaman tidak ditemukan.</p>
                  <a href="/" class="btn btn-primary mt-6">Kembali ke Beranda</a>
                </div>
              </body>
            </html>`)
	})

	host := env.Get("FIBER_HOST", "127.0.0.1")
	port := env.Get("FIBER_PORT", "8080")
	addr := fmt.Sprintf("%s:%s", host, port)

	log.Printf("Gothic Forge listening at http://%s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
