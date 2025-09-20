package routes

import (
	"github.com/gofiber/fiber/v2"
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
