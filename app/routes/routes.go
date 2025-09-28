package routes

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"gothicforge/app/ssg"
	"gothicforge/app/templates"
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
					if n, err := strconv.Atoi(t); err == nil {
						count = n
					}
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

	// Serve favicon for user agents that auto-request /favicon.ico
	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		// Prefer redirect to the SVG icon served under /static
		return c.Redirect("/static/favicon.svg", fiber.StatusMovedPermanently)
	})

	// Robots.txt and sitemap.xml: prefer static assets; fall back to generated content
	app.Get("/robots.txt", func(c *fiber.Ctx) error {
		p := filepath.Join("app", "static", "robots.txt")
		if _, err := os.Stat(p); err == nil {
			return c.Redirect("/static/robots.txt", fiber.StatusMovedPermanently)
		}
		// Dynamic, minimal robots.txt with optional Sitemap line
		siteMapURL := templates.ResolveCanonical("/sitemap.xml")
		var b strings.Builder
		b.WriteString("User-agent: *\n")
		b.WriteString("Allow: /\n")
		if siteMapURL != "" {
			b.WriteString("Sitemap: ")
			b.WriteString(siteMapURL)
			b.WriteString("\n")
		}
		c.Type("txt")
		return c.SendString(b.String())
	})
	app.Get("/sitemap.xml", func(c *fiber.Ctx) error {
		p := filepath.Join("app", "static", "sitemap.xml")
		if _, err := os.Stat(p); err == nil {
			return c.Redirect("/static/sitemap.xml", fiber.StatusMovedPermanently)
		}
		// Dynamic minimal sitemap from registered SSG pages
		pages := ssg.Pages()
		var b strings.Builder
		b.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
		b.WriteString("<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n")
		seen := make(map[string]struct{})
		// Always include root
		addURL := func(path string) {
			if _, ok := seen[path]; ok {
				return
			}
			seen[path] = struct{}{}
			loc := templates.ResolveCanonical(path)
			if loc == "" {
				loc = path
			}
			b.WriteString("  <url><loc>")
			b.WriteString(loc)
			b.WriteString("</loc></url>\n")
		}
		addURL("/")
		for _, p := range pages {
			if strings.TrimSpace(p.Path) == "" {
				continue
			}
			addURL(p.Path)
		}
		b.WriteString("</urlset>\n")
		c.Type("xml")
		return c.SendString(b.String())
	})

	// Apply any additional registrars contributed by scaffolds/features.
	applyRegistrars(app)
}
