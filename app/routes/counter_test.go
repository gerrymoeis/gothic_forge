package routes

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"gothicforge/internal/server"
)

func setupCounterApp() *fiber.App {
	app := server.New()
	Register(app)
	return app
}

func extractCookiePair(setCookie string, name string) string {
	// returns name=value from Set-Cookie header
	parts := strings.Split(setCookie, ";")
	if len(parts) > 0 && strings.HasPrefix(strings.TrimSpace(parts[0]), name+"=") {
		return strings.TrimSpace(parts[0])
	}
	// search other cookies if multiple Set-Cookie headers are concatenated
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, name+"=") {
			// some segments may be attributes; guard by excluding attribute keys like Path, Expires, etc.
			if strings.Contains(p, "=") && !strings.ContainsAny(p, " ") && !strings.HasPrefix(p, "Path=") {
				return p
			}
		}
	}
	return ""
}

func TestCounterPage_GET(t *testing.T) {
	app := setupCounterApp()
	req := httptest.NewRequest("GET", "/counter", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("/counter GET failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("expected text/html, got %q", ct)
	}
	body, _ := io.ReadAll(resp.Body)
	s := string(body)
	if !strings.Contains(s, "HTMX Counter") || !strings.Contains(s, "id=\"count-value\"") {
		t.Fatalf("expected counter page content, got: %s", s)
	}
}

func TestCounterIncrement_POST_Once(t *testing.T) {
	app := setupCounterApp()
	// Seed CSRF cookie via GET /counter
	get := httptest.NewRequest("GET", "/counter", nil)
	got, err := app.Test(get)
	if err != nil {
		t.Fatalf("seed GET failed: %v", err)
	}
	setCookie := got.Header.Get("Set-Cookie")
	csrfPair := extractCookiePair(setCookie, "_gforge_csrf")
	sessPair := extractCookiePair(setCookie, "session") // fiber session cookie name default is "session"
	if csrfPair == "" {
		t.Fatalf("missing csrf cookie in response")
	}

	req := httptest.NewRequest("POST", "/counter/increment", nil)
	// CSRF: header + cookie must match cookie token
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
	s := string(body)
	if !strings.Contains(s, "id=\"count-value\"") {
		t.Fatalf("expected counter value response, got: %s", s)
	}
}

func TestCounterIncrement_POST_TwiceWithSession(t *testing.T) {
	app := setupCounterApp()
	// Seed cookies via GET /counter
	g := httptest.NewRequest("GET", "/counter", nil)
	gr, err := app.Test(g)
	if err != nil {
		t.Fatalf("seed GET failed: %v", err)
	}
	sc := gr.Header.Get("Set-Cookie")
	csrfPair := extractCookiePair(sc, "_gforge_csrf")
	sessPair := extractCookiePair(sc, "session")
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
	// Capture session cookie
	sessCookie := resp1.Header.Get("Set-Cookie")
	// Second increment with session + csrf
	req2 := httptest.NewRequest("POST", "/counter/increment", nil)
	req2.Header.Set("X-CSRF-Token", token)
	// merge session cookie (may be refreshed) with csrfPair
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
	s2 := string(body2)
	if !strings.Contains(s2, `id="count-value"`) {
		t.Fatalf("expected counter value response, got: %s", s2)
	}
}
