package routes

import (
	"github.com/gofiber/fiber/v2"
	"gothicforge/app/db"
	"gothicforge/app/templates"
)

// Register mounts all application routes.
func Register(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		c.Type("html")
		return templates.Index().Render(c.UserContext(), c.Response().BodyWriter())
	})

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
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

	// Serve robots.txt and sitemap.xml from /static
	app.Get("/robots.txt", func(c *fiber.Ctx) error {
		return c.SendFile("./app/static/robots.txt", true)
	})
	app.Get("/sitemap.xml", func(c *fiber.Ctx) error {
		return c.SendFile("./app/static/sitemap.xml", true)
	})
}
