package routes

import (
  "bytes"

  "github.com/a-h/templ"
  "github.com/gofiber/fiber/v2"
  // removed unused templates import; routes are defined in hello.go
)

// render is a small helper to render a templ.Component into Fiber's response.
func render(c *fiber.Ctx, cmp templ.Component) error {
  var buf bytes.Buffer
  if err := cmp.Render(c.UserContext(), &buf); err != nil {
    return c.Status(fiber.StatusInternalServerError).SendString("template render error")
  }
  c.Type("html", "utf-8")
  return c.SendString(buf.String())
}

// NOTE: Routes are registered in hello.go. This file only provides the render helper.
