package routes

import (
    "context"
    "crypto/tls"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "os"
    "path/filepath"
    "sort"
    "strconv"
    "strings"
    "time"

    "github.com/go-chi/chi/v5"
    "gothicforge3/app/templates"
    redigo "github.com/gomodule/redigo/redis"
    "gothicforge3/internal/db"
    "gothicforge3/internal/env"
    "gothicforge3/internal/server"
    "gothicforge3/internal/auth"
)

// Register mounts all application routes on a chi router.
func Register(r *chi.Mux) {
    // Home
    r.Get("/", func(w http.ResponseWriter, req *http.Request) {
        server.Sessions().Put(req.Context(), "count", 0)
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        _ = templates.Index().Render(req.Context(), w)
    })

    // dev-only: mint a short-lived JWT and set cookie (gf_jwt)
    r.Get("/dev/jwt", func(w http.ResponseWriter, req *http.Request) {
        if strings.EqualFold(env.Get("APP_ENV", "development"), "production") {
            http.NotFound(w, req)
            return
        }
        sub := strings.TrimSpace(req.URL.Query().Get("sub"))
        if sub == "" { sub = "dev" }
        tok, exp, err := auth.Issue(1*time.Hour, map[string]any{"sub": sub, "role": "dev"})
        if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
        auth.SetJWTCookie(w, "gf_jwt", tok, exp)
        w.Header().Set("Content-Type", "text/plain; charset=utf-8")
        _, _ = w.Write([]byte("ok\n"))
    })

    // Register root for sitemap
    RegisterURL("/")

    // Counter sync (HTMX): accepts a count and returns the server stat fragment
    r.Post("/counter/sync", func(w http.ResponseWriter, req *http.Request) {
        if err := req.ParseForm(); err != nil {
            http.Error(w, "bad request", http.StatusBadRequest)
            return
        }
        v := strings.TrimSpace(req.FormValue("count"))
        n, err := strconv.Atoi(v)
        if err != nil { n = 0 }
        w.Header().Set("Content-Type", "text/plain; charset=utf-8")
        _, _ = w.Write([]byte(strconv.Itoa(n)))
    })

    // favicon redirect
    r.Get("/favicon.ico", func(w http.ResponseWriter, req *http.Request) {
        http.Redirect(w, req, "/static/favicon.svg", http.StatusMovedPermanently)
    })

    // health
    r.Get("/healthz", func(w http.ResponseWriter, req *http.Request) {
        w.Header().Set("Content-Type", "text/plain; charset=utf-8")
        _, _ = w.Write([]byte("ok"))
    })

    // readiness (checks external deps when configured)
    r.Get("/readyz", func(w http.ResponseWriter, req *http.Request) {
        w.Header().Set("Content-Type", "text/plain; charset=utf-8")
        status := http.StatusOK
        // Valkey readiness (optional)
        if err := valkeyPing(); err != nil {
            status = http.StatusServiceUnavailable
            _, _ = w.Write([]byte("valkey: FAIL\n"))
        } else {
            _, _ = w.Write([]byte("valkey: OK|SKIP\n"))
        }
        // DB readiness (optional)
        if skip, err := dbReady(); skip {
            _, _ = w.Write([]byte("db: SKIP\n"))
        } else if err != nil {
            status = http.StatusServiceUnavailable
            _, _ = w.Write([]byte("db: FAIL\n"))
        } else {
            _, _ = w.Write([]byte("db: OK\n"))
        }
        if status != http.StatusOK { w.WriteHeader(status) }
    })

    // robots.txt (serve from root). If a file exists under app/static, stream it directly; otherwise emit sensible defaults.
    r.Get("/robots.txt", func(w http.ResponseWriter, req *http.Request) {
        p := filepath.Join("app", "static", "robots.txt")
        if f, err := os.Open(p); err == nil {
            defer f.Close()
            w.Header().Set("Content-Type", "text/plain; charset=utf-8")
            _, _ = io.Copy(w, f)
            return
        }
        var b strings.Builder
        b.WriteString("User-agent: *\n")
        b.WriteString("Allow: /\n")
        b.WriteString("Sitemap: ")
        b.WriteString(absBaseURL(req))
        if !strings.HasSuffix(b.String(), "/") { b.WriteString("/") }
        b.WriteString("sitemap.xml\n")
        w.Header().Set("Content-Type", "text/plain; charset=utf-8")
        _, _ = w.Write([]byte(b.String()))
    })

    // sitemap.xml (serve from root). If a file exists, stream it; else emit a minimal but valid sitemap with absolute URLs.
    r.Get("/sitemap.xml", func(w http.ResponseWriter, req *http.Request) {
        p := filepath.Join("app", "static", "sitemap.xml")
        if f, err := os.Open(p); err == nil {
            defer f.Close()
            w.Header().Set("Content-Type", "application/xml; charset=utf-8")
            _, _ = io.Copy(w, f)
            return
        }
        base := absBaseURL(req)
        if !strings.HasSuffix(base, "/") { base += "/" }
        var b strings.Builder
        b.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
        b.WriteString("<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n")
        // Include registered URLs (includes "/"). With optional metadata.
        infos := ListURLInfo()
        sort.Slice(infos, func(i, j int) bool { return infos[i].Path < infos[j].Path })
        today := time.Now().UTC().Format("2006-01-02")
        for _, inf := range infos {
            pth := inf.Path
            u := base
            if strings.HasPrefix(pth, "/") { u += strings.TrimPrefix(pth, "/") } else { u += pth }
            _, _ = fmt.Fprintf(&b, "  <url><loc>%s</loc>", u)
            // Defaults if not provided
            lm := inf.LastMod
            if lm == "" { lm = today }
            cf := inf.ChangeFreq
            if cf == "" { cf = "weekly" }
            pr := inf.Priority
            if pr == "" {
                if pth == "/" { pr = "1.0" } else { pr = "0.7" }
            }
            _, _ = fmt.Fprintf(&b, "<lastmod>%s</lastmod>", lm)
            _, _ = fmt.Fprintf(&b, "<changefreq>%s</changefreq>", cf)
            _, _ = fmt.Fprintf(&b, "<priority>%s</priority>", pr)
            b.WriteString("</url>\n")
        }
        b.WriteString("</urlset>\n")
        w.Header().Set("Content-Type", "application/xml; charset=utf-8")
        _, _ = w.Write([]byte(b.String()))
    })

    // apply additional registrars
    applyRegistrars(r)
}

// absBaseURL returns SITE_BASE_URL if provided (normalized), otherwise derives from request scheme/host.
func absBaseURL(req *http.Request) string {
    if v := strings.TrimSpace(env.Get("SITE_BASE_URL", "")); v != "" {
        if strings.HasSuffix(v, "/") { return strings.TrimRight(v, "/") }
        return v
    }
    scheme := "http"
    if req.TLS != nil || strings.EqualFold(req.Header.Get("X-Forwarded-Proto"), "https") { scheme = "https" }
    host := req.Host
    return scheme + "://" + host
}

// dbReady tries to connect and ping the Postgres database when DATABASE_URL is set.
// Returns (skip=true) when DATABASE_URL is empty.
func dbReady() (bool, error) {
    if strings.TrimSpace(env.Get("DATABASE_URL", "")) == "" {
        return true, nil
    }
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
    if err := db.Connect(ctx); err != nil { return false, err }
    return false, db.Health(ctx)
}

// valkeyPing attempts to PING a Valkey/Redis instance when configured.
// It returns nil if VALKEY_URL/REDIS_URL is empty (treated as SKIP) or if PING succeeds.
// It returns an error only when a URL is configured but PING fails.
func valkeyPing() error {
    ru := strings.TrimSpace(env.Get("VALKEY_URL", ""))
    if ru == "" { ru = strings.TrimSpace(env.Get("REDIS_URL", "")) }
    if ru == "" { return nil } // not configured → skip is OK
    skipVerify := strings.EqualFold(strings.TrimSpace(env.Get("VALKEY_TLS_SKIP_VERIFY", "")), "1")
    u, perr := url.Parse(ru)
    var c redigo.Conn
    var err error
    if perr == nil {
        scheme := strings.ToLower(u.Scheme)
        if scheme == "rediss" || skipVerify {
            opts := []redigo.DialOption{}
            if u.User != nil {
                if pw, ok := u.User.Password(); ok { opts = append(opts, redigo.DialPassword(pw)) }
            }
            if dbStr := strings.TrimPrefix(u.Path, "/"); dbStr != "" {
                if n, e := strconv.Atoi(dbStr); e == nil { opts = append(opts, redigo.DialDatabase(n)) }
            }
            opts = append(opts, redigo.DialUseTLS(true))
            if skipVerify { opts = append(opts, redigo.DialTLSConfig(&tls.Config{InsecureSkipVerify: true})) }
            host := u.Host
            c, err = redigo.Dial("tcp", host, opts...)
        } else {
            c, err = redigo.DialURL(ru)
        }
    } else {
        c, err = redigo.DialURL(ru)
    }
    if err != nil { return err }
    defer c.Close()
    _, err = c.Do("PING")
    return err
}
