package routes

import (
	"github.com/gofiber/fiber/v2"
	"gothicforge/app/templates"
)

func init() {
	RegisterRoute(func(app *fiber.App) {
		app.Get("/about", func(c *fiber.Ctx) error {
			c.Type("html")
			return templates.About().Render(c.UserContext(), c.Response().BodyWriter())
		})
	})
}
