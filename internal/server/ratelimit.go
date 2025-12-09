package server

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/httprate"
	"gothicforge3/internal/env"
)

// RateLimitConfig holds configuration for intelligent rate limiting
type RateLimitConfig struct {
	// Enabled determines if rate limiting is active
	Enabled bool

	// PerIPLimit is the maximum requests per IP per window
	PerIPLimit int

	// GlobalLimit is the maximum total requests per window
	GlobalLimit int

	// Window is the time window for rate limiting
	Window time.Duration

	// ExemptPaths are paths that bypass rate limiting
	ExemptPaths []string

	// ExemptMethods are HTTP methods that bypass rate limiting
	ExemptMethods []string
}

// DefaultRateLimitConfig returns sensible default rate limit settings
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Enabled:     true,
		PerIPLimit:  100,              // 100 requests per minute per IP
		GlobalLimit: 1000,             // 1000 requests per minute globally
		Window:      time.Minute,      // 1 minute window
		ExemptPaths: []string{
			"/static/",
			"/favicon.ico",
			"/robots.txt",
			"/sitemap.xml",
			"/healthz",
			"/health",
			"/ping",
		},
		ExemptMethods: []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodOptions,
		},
	}
}

// LoadRateLimitConfig loads rate limit configuration from environment variables
func LoadRateLimitConfig() RateLimitConfig {
	config := DefaultRateLimitConfig()

	// Check if rate limiting is disabled
	if strings.EqualFold(strings.TrimSpace(env.Get("RATE_LIMIT_ENABLED", "true")), "false") {
		config.Enabled = false
		return config
	}

	// Per-IP limit
	if v := strings.TrimSpace(env.Get("RATE_LIMIT_PER_IP", "")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			config.PerIPLimit = n
		}
	}

	// Global limit
	if v := strings.TrimSpace(env.Get("RATE_LIMIT_GLOBAL", "")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			config.GlobalLimit = n
		}
	}

	// Window duration (in seconds)
	if v := strings.TrimSpace(env.Get("RATE_LIMIT_WINDOW_SECONDS", "")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			config.Window = time.Duration(n) * time.Second
		}
	}

	// Additional exempt paths (comma-separated)
	if v := strings.TrimSpace(env.Get("RATE_LIMIT_EXEMPT_PATHS", "")); v != "" {
		paths := strings.Split(v, ",")
		for _, p := range paths {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				config.ExemptPaths = append(config.ExemptPaths, trimmed)
			}
		}
	}

	return config
}

// IntelligentRateLimiter creates a rate limiting middleware with smart exemptions
// This middleware applies both per-IP and global rate limits while exempting safe paths
func IntelligentRateLimiter(config RateLimitConfig) func(http.Handler) http.Handler {
	if !config.Enabled {
		// Return a no-op middleware if rate limiting is disabled
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	// Create per-IP rate limiter
	perIPLimiter := httprate.LimitByIP(config.PerIPLimit, config.Window)

	// Create global rate limiter
	globalLimiter := httprate.Limit(config.GlobalLimit, config.Window)

	log.Printf("Rate limiting enabled: %d req/IP, %d req/global, window=%v",
		config.PerIPLimit, config.GlobalLimit, config.Window)

	return func(next http.Handler) http.Handler {
		// Apply both limiters in sequence
		limited := perIPLimiter(globalLimiter(next))

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if path should be exempted
			if shouldExemptFromRateLimit(r, config) {
				next.ServeHTTP(w, r)
				return
			}

			// Apply rate limiting
			limited.ServeHTTP(w, r)
		})
	}
}

// shouldExemptFromRateLimit determines if a request should bypass rate limiting
func shouldExemptFromRateLimit(r *http.Request, config RateLimitConfig) bool {
	path := r.URL.Path
	method := r.Method

	// Check exempt methods
	for _, exemptMethod := range config.ExemptMethods {
		if method == exemptMethod {
			return true
		}
	}

	// Check exempt paths
	for _, exemptPath := range config.ExemptPaths {
		if strings.HasPrefix(path, exemptPath) || path == strings.TrimSuffix(exemptPath, "/") {
			return true
		}
	}

	return false
}

// RateLimitStats provides statistics about rate limiting
type RateLimitStats struct {
	Enabled       bool
	PerIPLimit    int
	GlobalLimit   int
	Window        time.Duration
	ExemptPaths   []string
	ExemptMethods []string
}

// GetRateLimitStats returns current rate limit configuration for monitoring
func GetRateLimitStats(config RateLimitConfig) RateLimitStats {
	return RateLimitStats{
		Enabled:       config.Enabled,
		PerIPLimit:    config.PerIPLimit,
		GlobalLimit:   config.GlobalLimit,
		Window:        config.Window,
		ExemptPaths:   config.ExemptPaths,
		ExemptMethods: config.ExemptMethods,
	}
}
