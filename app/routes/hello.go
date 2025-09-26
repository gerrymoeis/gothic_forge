package routes

import (
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/session"
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

    // Health page removed for ultra-minimal starter.

    // Removed /healthz to keep the starter ultra-minimal.

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
                    if n, err := strconv.Atoi(t); err == nil {
                        count = n
                    }
                }
            } else {
                s.Set("count", 0)
                _ = s.Save()
            }
        }
        return templates.CounterWidget(count).Render(c.UserContext(), c.Response().BodyWriter())
    })
    // Removed full /counter page; home embeds the widget via HTMX.
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

    // Removed /db/ping endpoint from the starter surface.

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
