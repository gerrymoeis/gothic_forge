package routes

import (
    "github.com/gofiber/fiber/v2"
)

func init() {
    RegisterRoute(func(app *fiber.App) {
        app.Get("/api/example", func(c *fiber.Ctx) error {
            return c.JSON(fiber.Map{
                "ok": true,
                "data": fiber.Map{
                    "message": "Example endpoint",
                },
            })
        })
    })
}
