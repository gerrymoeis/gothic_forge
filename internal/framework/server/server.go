package server

import (
	"time"
	"os"
	"path/filepath"

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

	// Serve static assets from project root /app/static at /static
	app.Static("/static", detectStaticDir())

	return app
}

// detectStaticDir tries to locate the module root (folder containing go.mod)
// and returns an absolute path to app/static. Falls back to relative path.
func detectStaticDir() string {
    wd, _ := os.Getwd()
    cur := wd
    for {
        if _, err := os.Stat(filepath.Join(cur, "go.mod")); err == nil {
            p := filepath.Join(cur, "app", "static")
            if _, err := os.Stat(p); err == nil {
                return p
            }
            break
        }
        parent := filepath.Dir(cur)
        if parent == cur {
            break
        }
        cur = parent
    }
    // Fallback (works when running from repo root)
    return "./app/static"
}
