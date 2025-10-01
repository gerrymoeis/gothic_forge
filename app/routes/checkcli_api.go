package routes

import (
	"github.com/gofiber/fiber/v2"
)

func init() {
	RegisterRoute(func(app *fiber.App) {
		app.Get("/api/checkcli", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{
				"ok": true,
				"data": fiber.Map{
					"message": "Checkcli endpoint",
				},
			})
		})
	})
}
