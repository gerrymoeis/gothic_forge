package server

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"gothicforge/internal/env"

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
		Prefork:           false,
		AppName:           "Gothic Forge",
	})

	// Core middlewares
	app.Use(requestid.New())
	app.Use(recover.New())
	// Logger: plain or JSON depending on LOG_FORMAT
	if env.Get("LOG_FORMAT", "") == "json" {
		app.Use(logger.New(logger.Config{
			Format:     `{"time":"${time}","id":"${locals:requestid}","ip":"${ip}","method":"${method}","path":"${path}","status":${status},"latency":"${latency}","ua":"${ua}"}\n`,
			TimeFormat: time.RFC3339,
		}))
	} else {
		app.Use(logger.New())
	}
	app.Use(helmet.New())
	// Some Helmet defaults can interfere with cross-origin resources (e.g., COEP/COOP).
	// Remove them to allow CDN CSS/JS while we rely on strict CSP for safety.
	app.Use(func(c *fiber.Ctx) error {
		if err := c.Next(); err != nil {
			return err
		}
		// Operate directly on the response header to avoid copylocks and ensure deletions apply.
		c.Response().Header.Del("Cross-Origin-Embedder-Policy")
		c.Response().Header.Del("Cross-Origin-Opener-Policy")
		// Cross-Origin-Resource-Policy applies to resources we serve; not needed for our pages.
		// Leaving it unset avoids blocking external CDN resources unintentionally.
		c.Response().Header.Del("Cross-Origin-Resource-Policy")
		return nil
	})
	app.Use(compress.New())
	// CORS: allow configurable origins via CORS_ORIGINS (comma-separated). Defaults to permissive.
	if origins := env.Get("CORS_ORIGINS", ""); origins != "" {
		o := strings.TrimSpace(origins)
		cfg := cors.Config{
			AllowOrigins: o,
			AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Requested-With, X-CSRF-Token",
			AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		}
		// AllowCredentials must not be true when origins is '*', per Fiber CORS rules.
		if o != "*" {
			cfg.AllowCredentials = true
		}
		app.Use(cors.New(cfg))
	} else {
		app.Use(cors.New())
	}

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
	app.Use(csrf.New(csrf.Config{
		KeyLookup:      "header:X-CSRF-Token",
		CookieName:     "_gforge_csrf",
		CookieHTTPOnly: false, // allow HTMX helper to read cookie and set header
		CookieSecure:   env.Get("APP_ENV", "development") == "production",
		CookieSameSite: "Lax",
		Expiration:     12 * time.Hour,
	}))

	// Content Security Policy (CSP)
	app.Use(func(c *fiber.Ctx) error {
		// Remove any existing CSP header (e.g., from Helmet defaults) to avoid multiple CSP headers
		// which browsers treat cumulatively (intersection), potentially blocking CDN resources.
		c.Response().Header.Del("Content-Security-Policy")
		// In development, allow any https for script/style to avoid friction with CDNs.
		// In production, keep a strict allowlist.
		var csp []string
		if env.Get("APP_ENV", "development") == "development" {
			csp = []string{
				"default-src 'self'",
				"script-src 'self' https:",
				"style-src 'self' https: 'unsafe-inline'",
				"img-src 'self' data: https:",
				"font-src 'self' https:",
				"connect-src 'self' https:",
				"object-src 'none'",
				"base-uri 'self'",
				"frame-ancestors 'self'",
			}
		} else {
			csp = []string{
				"default-src 'self'",
				"script-src 'self' https://unpkg.com",
				"style-src 'self' https://cdn.jsdelivr.net",
				"img-src 'self' data: https:",
				"font-src 'self' https://cdn.jsdelivr.net",
				"connect-src 'self'",
				"object-src 'none'",
				"base-uri 'self'",
				"frame-ancestors 'self'",
			}
		}
		c.Set("Content-Security-Policy", strings.Join(csp, "; "))
		return c.Next()
	})

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
