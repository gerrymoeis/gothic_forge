package server

import (
	"time"

	"gothicforge/internal/framework/env"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// New constructs a Fiber app with secure, production-ready defaults.
func New() *fiber.App {
	app := fiber.New(fiber.Config{
		EnablePrintRoutes: env.Get("APP_ENV", "development") == "development",
		Prefork:          false,
		AppName:          "Gothic Forge",
	})

	// Core middlewares
	app.Use(requestid.New())
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(helmet.New())
	app.Use(compress.New())
	app.Use(cors.New())

	// Rate limiter (sane defaults)
	app.Use(limiter.New(limiter.Config{
		Max:        120,
		Expiration: 1 * time.Minute,
	}))

	// Session store (cookie-based, in-memory store)
	sess := session.New(session.Config{
		CookieHTTPOnly: true,
		CookieSecure:   env.Get("APP_ENV", "development") == "production",
		CookieSameSite: "Lax",
	})
	app.Use(func(c *fiber.Ctx) error {
		// attach session to context for handlers to use via Locals
		s, err := sess.Get(c)
		if err != nil {
			return err
		}
		c.Locals("session", s)
		return c.Next()
	})

	// CSRF protection (header-based token)
	csrfSecret := env.Get("CSRF_SECRET", "random-secret")
	app.Use(csrf.New(csrf.Config{
		KeyLookup:     "header:X-CSRF-Token",
		CookieName:    "_gforge_csrf",
		CookieSecure:  env.Get("APP_ENV", "development") == "production",
		CookieSameSite: "Lax",
		Expiration:    12 * time.Hour,
		KeyGenerator:  func() string { return csrfSecret },
	}))

	// Serve static assets from /app/static at /static
	app.Static("/static", "./app/static")

	return app
}
