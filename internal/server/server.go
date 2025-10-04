package server

import (
    "fmt"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "time"

    "github.com/alexedwards/scs/v2"
    "github.com/alexedwards/scs/redisstore"
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/go-chi/cors"
    "github.com/go-chi/httprate"
    redigo "github.com/gomodule/redigo/redis"
    "gothicforge3/internal/env"
)

var sessionManager *scs.SessionManager
// Sessions exposes the global session manager
func Sessions() *scs.SessionManager { return sessionManager }
// New creates the chi router with defaults
func New() *chi.Mux {
    r := chi.NewRouter()
    // Core middlewares
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Recoverer)
    switch strings.ToLower(env.Get("LOG_FORMAT", "")) {
    case "json":
        r.Use(jsonLoggerMiddleware)
    case "off", "silent", "none":
        // no request logging middleware
    default:
        r.Use(middleware.Logger)
    }
    r.Use(middleware.Compress(5))

    // CORS
    r.Use(configureCORS())

    // Rate limit (bypass safe paths)
    maxReq := 120
    if v := strings.TrimSpace(env.Get("RATE_LIMIT_MAX", "")); v != "" {
        if n, err := strconv.Atoi(v); err == nil && n > 0 {
            maxReq = n
        }
    }
    window := time.Minute
    if v := strings.TrimSpace(env.Get("RATE_LIMIT_WINDOW_SECONDS", "")); v != "" {
        if n, err := strconv.Atoi(v); err == nil && n > 0 {
            window = time.Duration(n) * time.Second
        }
    }
    lim := httprate.LimitByIP(maxReq, window)
    r.Use(func(next http.Handler) http.Handler {
        limited := lim(next)
        return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
            p := req.URL.Path
            if req.Method == http.MethodGet || req.Method == http.MethodHead || req.Method == http.MethodOptions || strings.HasPrefix(p, "/static/") || p == "/favicon.ico" || p == "/robots.txt" || p == "/sitemap.xml" || p == "/healthz" {
                next.ServeHTTP(w, req)
                return
            }
            limited.ServeHTTP(w, req)
        })
    })

    // Sessions (cookie-based)
    sessionManager = scs.New()
    // 24h lifetime by default
    sessionManager.Lifetime = 24 * time.Hour
    sessionManager.Cookie.HttpOnly = true
    sessionManager.Cookie.SameSite = http.SameSiteLaxMode
    sessionManager.Cookie.Secure = env.Get("APP_ENV", "development") == "production"
    // Valkey/Redis session store if URL provided (redigo pool)
    ru := strings.TrimSpace(env.Get("VALKEY_URL", ""))
    if ru == "" {
        ru = strings.TrimSpace(env.Get("REDIS_URL", ""))
    }
    if ru != "" {
        pool := &redigo.Pool{
            MaxIdle:     4,
            IdleTimeout: 300 * time.Second,
            Dial: func() (redigo.Conn, error) {
                return redigo.DialURL(ru)
            },
            TestOnBorrow: func(c redigo.Conn, t time.Time) error {
                if time.Since(t) < time.Minute { return nil }
                _, err := c.Do("PING")
                return err
            },
        }
        sessionManager.Store = redisstore.New(pool)
    }
    r.Use(sessionManager.LoadAndSave)

    // Content-Security-Policy
    r.Use(func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
            var csp []string
            if env.Get("APP_ENV", "development") == "development" {
                csp = []string{
                    "default-src 'self'",
                    "script-src 'self' https: 'unsafe-eval' 'unsafe-inline'",
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
                    // Allow CDNs used in Layout/SEO and inline JSON-LD
                    "script-src 'self' https://unpkg.com https://cdn.jsdelivr.net 'unsafe-inline'",
                    "style-src 'self' https: 'unsafe-inline'",
                    "img-src 'self' data: https:",
                    "font-src 'self' https:",
                    "connect-src 'self' https:",
                    "object-src 'none'",
                    "base-uri 'self'",
                    "frame-ancestors 'self'",
                }
            }
            w.Header().Set("Content-Security-Policy", strings.Join(csp, "; "))
            next.ServeHTTP(w, req)
        })
    })

    // CSRF (prod only): simple same-origin check for state-changing requests
    if env.Get("APP_ENV", "development") == "production" {
        r.Use(CSRFMiddleware)
    }

    // Static assets (CSS/JS/images)
    mountStatic(r)

    return r
}

func mountStatic(r *chi.Mux) {
    // serve app/static under /static
    staticDir := detectStaticDir()
    fs := http.StripPrefix("/static", http.FileServer(http.Dir(staticDir)))
    r.Handle("/static/*", fs)
    // also serve styles directly for convenience
    styles := http.StripPrefix("/static/styles", http.FileServer(http.Dir(filepath.Join("app", "styles"))))
    r.Handle("/static/styles/*", styles)
}

func detectStaticDir() string {
    if base := strings.TrimSpace(env.Get("GFORGE_BASEDIR", "")); base != "" {
        return filepath.Join(base, "app", "static")
    }
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
        if parent == cur { break }
        cur = parent
    }
    return filepath.Join("app", "static")
}

func configureCORS() func(http.Handler) http.Handler {
	origins := strings.TrimSpace(env.Get("CORS_ORIGINS", ""))
	opts := cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}
	if origins == "" {
		opts.AllowedOrigins = []string{"*"}
	} else {
		parts := strings.Split(origins, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" { out = append(out, p) }
		}
		if len(out) == 0 { out = []string{"*"} }
		opts.AllowedOrigins = out
		allowCreds := true
		for _, o := range out { if o == "*" { allowCreds = false; break } }
		opts.AllowCredentials = allowCreds
	}
	return cors.Handler(opts)
}

func jsonLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		entry := fmt.Sprintf(`{"time":"%s","ip":"%s","method":"%s","path":"%s","status":%d,"bytes":%d,"latency":"%s","ua":"%s"}`,
			time.Now().Format(time.RFC3339), r.RemoteAddr, r.Method, r.URL.Path, ww.Status(), ww.BytesWritten(), time.Since(start).String(), r.UserAgent())
		log.Println(entry)
	})
}
