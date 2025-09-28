package routes

import (
	"context"
	"net/mail"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"gothicforge/app/db"
	"gothicforge/app/templates"
	"gothicforge/security"
)

func init() {
	RegisterRoute(func(app *fiber.App) {
		// Aliases for convenience
		app.Get("/login", func(c *fiber.Ctx) error {
			return c.Redirect("/auth/login", fiber.StatusMovedPermanently)
		})
		app.Get("/register", func(c *fiber.Ctx) error {
			return c.Redirect("/auth/register", fiber.StatusMovedPermanently)
		})

		// GET forms
		app.Get("/auth/register", func(c *fiber.Ctx) error {
			c.Type("html")
			return templates.AuthRegister().Render(c.UserContext(), c.Response().BodyWriter())
		})
		app.Get("/auth/login", func(c *fiber.Ctx) error {
			c.Type("html")
			return templates.AuthLogin().Render(c.UserContext(), c.Response().BodyWriter())
		})

		// GET logout confirmation (POST performs actual logout to remain CSRF-safe)
		app.Get("/auth/logout", func(c *fiber.Ctx) error {
			c.Type("html")
			return templates.AuthLogout().Render(c.UserContext(), c.Response().BodyWriter())
		})

		// POST register
		app.Post("/auth/register", func(c *fiber.Ctx) error {
			ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
			defer cancel()
			pool, _ := db.ConnectCached(ctx)
			if pool == nil {
				return c.Status(fiber.StatusServiceUnavailable).SendString("database not configured")
			}
			_ = db.EnsureUsersTable(ctx, pool)
			email := strings.TrimSpace(c.FormValue("email"))
			pass := c.FormValue("password")
			if email == "" || pass == "" {
				return c.Status(fiber.StatusBadRequest).SendString("email and password required")
			}
			if _, err := mail.ParseAddress(email); err != nil {
				return c.Status(fiber.StatusBadRequest).SendString("invalid email")
			}
			if len(pass) < 8 {
				return c.Status(fiber.StatusBadRequest).SendString("password too short")
			}
			hash, err := security.HashPassword(pass)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString("failed to hash password")
			}
			u, err := db.CreateUser(ctx, pool, email, hash)
			if err != nil {
				return c.Status(fiber.StatusConflict).SendString("could not create user")
			}
			if s, ok := c.Locals("session").(*session.Session); ok && s != nil {
				s.Set("user_id", u.ID.String())
				_ = s.Save()
			}
			c.Set("HX-Redirect", "/")
			return c.Redirect("/", fiber.StatusSeeOther)
		})

		// POST login
		app.Post("/auth/login", func(c *fiber.Ctx) error {
			ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
			defer cancel()
			pool, _ := db.ConnectCached(ctx)
			if pool == nil {
				return c.Status(fiber.StatusServiceUnavailable).SendString("database not configured")
			}
			email := strings.TrimSpace(c.FormValue("email"))
			pass := c.FormValue("password")
			if email == "" || pass == "" {
				return c.Status(fiber.StatusBadRequest).SendString("email and password required")
			}
			u, err := db.GetUserByEmail(ctx, pool, email)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).SendString("invalid credentials")
			}
			if ok, _ := security.VerifyPassword(pass, u.PasswordHash); !ok {
				return c.Status(fiber.StatusUnauthorized).SendString("invalid credentials")
			}
			if s, ok := c.Locals("session").(*session.Session); ok && s != nil {
				s.Set("user_id", u.ID.String())
				_ = s.Save()
			}
			c.Set("HX-Redirect", "/")
			return c.Redirect("/", fiber.StatusSeeOther)
		})

		// POST logout
		app.Post("/auth/logout", func(c *fiber.Ctx) error {
			if s, ok := c.Locals("session").(*session.Session); ok && s != nil {
				s.Delete("user_id")
				_ = s.Save()
			}
			c.Set("HX-Redirect", "/")
			return c.Redirect("/", fiber.StatusSeeOther)
		})

		// GET me
		app.Get("/auth/me", func(c *fiber.Ctx) error {
			var uid string
			if s, ok := c.Locals("session").(*session.Session); ok && s != nil {
				if v := s.Get("user_id"); v != nil {
					if sv, ok := v.(string); ok {
						uid = sv
					}
				}
			}
			if uid == "" {
				return c.SendStatus(fiber.StatusUnauthorized)
			}
			return c.JSON(fiber.Map{"id": uid})
		})
	})
}
