package routes

import (
	"github.com/gofiber/fiber/v2"
	"gothicforge/app/templates"
)

func init() {
	RegisterRoute(func(app *fiber.App) {
		app.Get("/sample-cli", func(c *fiber.Ctx) error {
			c.Type("html")
			return templates.SampleCli().Render(c.UserContext(), c.Response().BodyWriter())
		})
	})
}
