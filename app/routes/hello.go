package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"gothicforge/app/db"
	"gothicforge/app/templates"
	"strconv"
)

// Register mounts all application routes.
func Register(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		// Reset counter on each visit to the homepage for a predictable demo
		if s, ok := c.Locals("session").(*session.Session); ok && s != nil {
			s.Set("count", 0)
			_ = s.Save()
		}
		c.Type("html")
		return templates.Index().Render(c.UserContext(), c.Response().BodyWriter())
	})

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	// HTMX Counter demo (session-backed)
	app.Get("/counter/widget", func(c *fiber.Ctx) error {
		c.Type("html")
		var count int
		if s, ok := c.Locals("session").(*session.Session); ok && s != nil {
			if v := s.Get("count"); v != nil {
				switch t := v.(type) {
				case int:
					count = t
				case int64:
					count = int(t)
				case string:
					if n, err := strconv.Atoi(t); err == nil { count = n }
				}
			} else {
				s.Set("count", 0)
				_ = s.Save()
			}
		}
		return templates.CounterWidget(count).Render(c.UserContext(), c.Response().BodyWriter())
	})
	app.Get("/counter", func(c *fiber.Ctx) error {
		c.Type("html")
		// Always reset to 0 for a predictable demo; do not read back to avoid edge cases
		if s, ok := c.Locals("session").(*session.Session); ok && s != nil {
			s.Set("count", 0)
			_ = s.Save()
		}
		count := 0
		return templates.CounterPage(count).Render(c.UserContext(), c.Response().BodyWriter())
	})
	app.Post("/counter/increment", func(c *fiber.Ctx) error {
		// CSRF middleware expects X-CSRF-Token header; set by /static/app.js
		var count int
		if s, ok := c.Locals("session").(*session.Session); ok && s != nil {
			if v := s.Get("count"); v != nil {
				switch t := v.(type) {
				case int:
					count = t
				case int64:
					count = int(t)
				case string:
					if n, err := strconv.Atoi(t); err == nil { count = n }
				}
			}
			count++
			s.Set("count", count)
			_ = s.Save()
		} else {
			count = 1
		}
		// Respond with plain digits; htmx will swap into #count-value via innerHTML
		return c.SendString(strconv.Itoa(count))
	})

	// Database ping (non-fatal when not configured)
	app.Get("/db/ping", func(c *fiber.Ctx) error {
		ctx := c.UserContext()
		p, err := db.Connect(ctx)
		if err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"ok":      false,
				"message": err.Error(),
			})
		}
		if p == nil {
			return c.JSON(fiber.Map{
				"ok":      false,
				"message": "database not configured",
			})
		}
		return c.JSON(fiber.Map{"ok": true})
	})

	// Serve favicon for user agents that auto-request /favicon.ico
	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		// Prefer redirect to the SVG icon served under /static
		return c.Redirect("/static/favicon.svg", fiber.StatusMovedPermanently)
	})

	// Serve robots.txt and sitemap.xml via redirect to /static (consistent with favicon)
	app.Get("/robots.txt", func(c *fiber.Ctx) error {
		return c.Redirect("/static/robots.txt", fiber.StatusMovedPermanently)
	})
	app.Get("/sitemap.xml", func(c *fiber.Ctx) error {
		return c.Redirect("/static/sitemap.xml", fiber.StatusMovedPermanently)
	})
}
