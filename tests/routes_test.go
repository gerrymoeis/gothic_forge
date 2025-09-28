package tests

import (
    "io"
    "net/http/httptest"
    "strings"
    "testing"
    "unicode"

    "github.com/gofiber/fiber/v2"
    "gothicforge/app/routes"
    "gothicforge/internal/server"
)

func setupApp() *fiber.App {
    app := server.New()
    routes.Register(app)
    return app
}

func extractCookiePair(setCookie string, name string) string {
    parts := strings.Split(setCookie, ";")
    if len(parts) > 0 && strings.HasPrefix(strings.TrimSpace(parts[0]), name+"=") {
        return strings.TrimSpace(parts[0])
    }
    for _, p := range parts {
        p = strings.TrimSpace(p)
        if strings.HasPrefix(p, name+"=") {
            if strings.Contains(p, "=") && !strings.ContainsAny(p, " ") && !strings.HasPrefix(p, "Path=") {
                return p
            }
        }
    }
    return ""
}

// extractCookiePairFromValues scans multiple Set-Cookie header values and
// returns the first matching name=value pair if present.
func extractCookiePairFromValues(values []string, name string) string {
    for _, v := range values {
        if p := extractCookiePair(v, name); p != "" {
            return p
        }
    }
    return ""
}

func TestHome(t *testing.T) {
    app := setupApp()
    req := httptest.NewRequest("GET", "/", nil)
    resp, err := app.Test(req)
    if err != nil {
        t.Fatalf("home request failed: %v", err)
    }
    if resp.StatusCode != fiber.StatusOK {
        t.Fatalf("expected 200, got %d", resp.StatusCode)
    }
    ct := resp.Header.Get("Content-Type")
    if !strings.Contains(ct, "text/html") {
        t.Fatalf("expected text/html content-type, got %q", ct)
    }
}

func TestFaviconRedirect(t *testing.T) {
    app := setupApp()
    req := httptest.NewRequest("GET", "/favicon.ico", nil)
    resp, err := app.Test(req)
    if err != nil {
        t.Fatalf("favicon request failed: %v", err)
    }
    if resp.StatusCode != fiber.StatusMovedPermanently {
        t.Fatalf("expected 301, got %d", resp.StatusCode)
    }
    loc := resp.Header.Get("Location")
    if loc != "/static/favicon.svg" {
        t.Fatalf("expected redirect to /static/favicon.svg, got %q", loc)
    }
}

func TestSitemapRedirect(t *testing.T) {
    app := setupApp()
    req := httptest.NewRequest("GET", "/sitemap.xml", nil)
    resp, err := app.Test(req)
    if err != nil {
        t.Fatalf("sitemap request failed: %v", err)
    }
    switch resp.StatusCode {
    case fiber.StatusMovedPermanently:
        loc := resp.Header.Get("Location")
        if loc != "/static/sitemap.xml" {
            t.Fatalf("expected redirect to /static/sitemap.xml, got %q", loc)
        }
    case fiber.StatusOK:
        // Dynamic fallback served directly; ensure minimal XML structure.
        b, _ := io.ReadAll(resp.Body)
        s := strings.TrimSpace(string(b))
        if s == "" || !strings.Contains(s, "<urlset") {
            t.Fatalf("expected sitemap xml body, got: %q", s)
        }
    default:
        t.Fatalf("unexpected status %d", resp.StatusCode)
    }
}

func TestCounterIncrement_POST_Once(t *testing.T) {
    app := setupApp()
    // Seed cookies via GET /
    get := httptest.NewRequest("GET", "/", nil)
    got, err := app.Test(get)
    if err != nil {
        t.Fatalf("seed GET failed: %v", err)
    }
    sc := got.Header.Values("Set-Cookie")
    csrfPair := extractCookiePairFromValues(sc, "_gforge_csrf")
    sessPair := extractCookiePairFromValues(sc, "session")
    if csrfPair == "" {
        t.Fatalf("missing csrf cookie in response")
    }

    // Single increment
    req := httptest.NewRequest("POST", "/counter/increment", nil)
    token := strings.SplitN(csrfPair, "=", 2)[1]
    req.Header.Set("X-CSRF-Token", token)
    if sessPair != "" {
        req.Header.Set("Cookie", sessPair+"; "+csrfPair)
    } else {
        req.Header.Set("Cookie", csrfPair)
    }
    resp, err := app.Test(req)
    if err != nil {
        t.Fatalf("/counter/increment POST failed: %v", err)
    }
    if resp.StatusCode != fiber.StatusOK {
        t.Fatalf("expected 200, got %d", resp.StatusCode)
    }
    body, _ := io.ReadAll(resp.Body)
    s := strings.TrimSpace(string(body))
    if s == "" {
        t.Fatalf("empty response")
    }
    for _, r := range s {
        if !unicode.IsDigit(r) {
            t.Fatalf("expected digits, got: %q", s)
        }
    }
}

func TestCounterIncrement_POST_TwiceWithSession(t *testing.T) {
    app := setupApp()
    // Seed cookies via GET /
    g := httptest.NewRequest("GET", "/", nil)
    gr, err := app.Test(g)
    if err != nil {
        t.Fatalf("seed GET failed: %v", err)
    }
    sc := gr.Header.Values("Set-Cookie")
    csrfPair := extractCookiePairFromValues(sc, "_gforge_csrf")
    sessPair := extractCookiePairFromValues(sc, "session")
    if csrfPair == "" {
        t.Fatalf("missing csrf cookie in response")
    }

    // First increment
    req1 := httptest.NewRequest("POST", "/counter/increment", nil)
    token := strings.SplitN(csrfPair, "=", 2)[1]
    req1.Header.Set("X-CSRF-Token", token)
    if sessPair != "" {
        req1.Header.Set("Cookie", sessPair+"; "+csrfPair)
    } else {
        req1.Header.Set("Cookie", csrfPair)
    }
    resp1, err := app.Test(req1)
    if err != nil {
        t.Fatalf("first increment failed: %v", err)
    }

    // Capture (possibly refreshed) session cookie
    sessValues := resp1.Header.Values("Set-Cookie")
    sessCookie := extractCookiePairFromValues(sessValues, "session")

    // Second increment with session + csrf
    req2 := httptest.NewRequest("POST", "/counter/increment", nil)
    req2.Header.Set("X-CSRF-Token", token)
    if sessCookie != "" {
        semi := strings.Index(sessCookie, ";")
        if semi > 0 {
            sessCookie = sessCookie[:semi]
        }
        req2.Header.Set("Cookie", sessCookie+"; "+csrfPair)
    } else if sessPair != "" {
        req2.Header.Set("Cookie", sessPair+"; "+csrfPair)
    } else {
        req2.Header.Set("Cookie", csrfPair)
    }
    resp2, err := app.Test(req2)
    if err != nil {
        t.Fatalf("second increment failed: %v", err)
    }
    if resp2.StatusCode != fiber.StatusOK {
        t.Fatalf("expected 200, got %d", resp2.StatusCode)
    }
    body2, _ := io.ReadAll(resp2.Body)
    s2 := strings.TrimSpace(string(body2))
    if s2 == "" {
        t.Fatalf("empty response")
    }
    for _, r := range s2 {
        if !unicode.IsDigit(r) {
            t.Fatalf("expected digits, got: %q", s2)
        }
    }
}
