//go:build integration && authscaffold
// +build integration,authscaffold

package tests

import (
	"context"
	"io"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"gothicforge/app/db"
	"gothicforge/app/routes"
	"gothicforge/internal/env"
	"gothicforge/internal/server"
)

func hasDatabase() bool {
	if strings.TrimSpace(os.Getenv("DATABASE_URL")) == "" {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()
	pool, err := db.ConnectCached(ctx)
	if err != nil || pool == nil {
		return false
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel2()
	if err := pool.Ping(ctx2); err != nil {
		return false
	}
	return true
}

func TestAuth_Register_DuplicateEmail_Conflict_SkipIfNoDB(t *testing.T) {
	if !hasDatabase() {
		t.Skip("DATABASE_URL not set; skipping auth integration test")
	}
	app := buildAuthTestApp()
	// Seed CSRF
	g := httptest.NewRequest("GET", "/", nil)
	gr, err := app.Test(g)
	if err != nil {
		t.Fatalf("seed GET failed: %v", err)
	}
	csrfPair := extractCookiePairFromValues(gr.Header.Values("Set-Cookie"), "_gforge_csrf")
	if csrfPair == "" {
		t.Fatalf("missing csrf cookie")
	}
	csrfToken := strings.SplitN(csrfPair, "=", 2)[1]

	email := "dupe+test@example.com"
	pass := "supersecret"

	// First register
	f1 := url.Values{}
	f1.Set("email", email)
	f1.Set("password", pass)
	r1 := httptest.NewRequest("POST", "/auth/register", strings.NewReader(f1.Encode()))
	r1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r1.Header.Set("X-CSRF-Token", csrfToken)
	r1.Header.Set("Cookie", csrfPair)
	resp1, err := app.Test(r1)
	if err != nil {
		t.Fatalf("register #1 failed: %v", err)
	}
	if resp1.StatusCode != fiber.StatusSeeOther {
		b, _ := io.ReadAll(resp1.Body)
		t.Fatalf("expected 303, got %d (%s)", resp1.StatusCode, string(b))
	}

	// Second register with same email → 409
	f2 := url.Values{}
	f2.Set("email", email)
	f2.Set("password", pass)
	r2 := httptest.NewRequest("POST", "/auth/register", strings.NewReader(f2.Encode()))
	r2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r2.Header.Set("X-CSRF-Token", csrfToken)
	r2.Header.Set("Cookie", csrfPair)
	resp2, err := app.Test(r2)
	if err != nil {
		t.Fatalf("register #2 failed: %v", err)
	}
	if resp2.StatusCode != fiber.StatusConflict {
		b, _ := io.ReadAll(resp2.Body)
		t.Fatalf("expected 409, got %d (%s)", resp2.StatusCode, string(b))
	}
}

func TestAuth_Register_InvalidEmail_BadRequest_SkipIfNoDB(t *testing.T) {
	if !hasDatabase() {
		t.Skip("DATABASE_URL not set; skipping auth integration test")
	}
	app := buildAuthTestApp()
	// Seed CSRF
	g := httptest.NewRequest("GET", "/", nil)
	gr, err := app.Test(g)
	if err != nil {
		t.Fatalf("seed GET failed: %v", err)
	}
	csrfPair := extractCookiePairFromValues(gr.Header.Values("Set-Cookie"), "_gforge_csrf")
	if csrfPair == "" {
		t.Fatalf("missing csrf cookie")
	}
	csrfToken := strings.SplitN(csrfPair, "=", 2)[1]

	f := url.Values{}
	f.Set("email", "not-an-email")
	f.Set("password", "supersecret")
	r := httptest.NewRequest("POST", "/auth/register", strings.NewReader(f.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("X-CSRF-Token", csrfToken)
	r.Header.Set("Cookie", csrfPair)
	resp, err := app.Test(r)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 400, got %d (%s)", resp.StatusCode, string(b))
	}
}

func TestAuth_Register_ShortPassword_BadRequest_SkipIfNoDB(t *testing.T) {
	if !hasDatabase() {
		t.Skip("DATABASE_URL not set; skipping auth integration test")
	}
	app := buildAuthTestApp()
	// Seed CSRF
	g := httptest.NewRequest("GET", "/", nil)
	gr, err := app.Test(g)
	if err != nil {
		t.Fatalf("seed GET failed: %v", err)
	}
	csrfPair := extractCookiePairFromValues(gr.Header.Values("Set-Cookie"), "_gforge_csrf")
	if csrfPair == "" {
		t.Fatalf("missing csrf cookie")
	}
	csrfToken := strings.SplitN(csrfPair, "=", 2)[1]

	f := url.Values{}
	f.Set("email", "shortpass@example.com")
	f.Set("password", "1234567") // 7 chars
	r := httptest.NewRequest("POST", "/auth/register", strings.NewReader(f.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("X-CSRF-Token", csrfToken)
	r.Header.Set("Cookie", csrfPair)
	resp, err := app.Test(r)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 400, got %d (%s)", resp.StatusCode, string(b))
	}
}

func buildAuthTestApp() *fiber.App {
	_ = env.Load()
	app := server.New()
	routes.Register(app)
	return app
}

func TestAuth_Register_Me_Logout_Me_SkipIfNoDB(t *testing.T) {
	if !hasDatabase() {
		t.Skip("DATABASE_URL not set; skipping auth integration test")
	}
	app := buildAuthTestApp()

	g := httptest.NewRequest("GET", "/", nil)
	gr, err := app.Test(g)
	if err != nil {
		t.Fatalf("seed GET failed: %v", err)
	}
	sc := gr.Header.Values("Set-Cookie")
	csrfPair := extractCookiePairFromValues(sc, "_gforge_csrf")
	if csrfPair == "" {
		t.Fatalf("missing csrf cookie in response")
	}
	csrfToken := strings.SplitN(csrfPair, "=", 2)[1]

	// Register
	form := url.Values{}
	form.Set("email", "user+test@example.com")
	form.Set("password", "supersecret")
	req := httptest.NewRequest("POST", "/auth/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-CSRF-Token", csrfToken)
	req.Header.Set("Cookie", csrfPair)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusSeeOther {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 303, got %d body=%s", resp.StatusCode, string(b))
	}
	// capture session cookie
	sessPair := extractCookiePairFromValues(resp.Header.Values("Set-Cookie"), "session")
	if sessPair == "" {
		// it may have been set earlier; try from seed response
		sessPair = extractCookiePairFromValues(sc, "session")
	}

	// GET /auth/me should be authorized now
	me := httptest.NewRequest("GET", "/auth/me", nil)
	if sessPair != "" {
		me.Header.Set("Cookie", sessPair)
	}
	mer, err := app.Test(me)
	if err != nil {
		t.Fatalf("/auth/me failed: %v", err)
	}
	if mer.StatusCode != fiber.StatusOK {
		b, _ := io.ReadAll(mer.Body)
		t.Fatalf("expected 200, got %d (%s)", mer.StatusCode, string(b))
	}

	// POST /auth/logout
	lg := httptest.NewRequest("POST", "/auth/logout", nil)
	lg.Header.Set("X-CSRF-Token", csrfToken)
	ck := csrfPair
	if sessPair != "" {
		ck = sessPair + "; " + csrfPair
	}
	lg.Header.Set("Cookie", ck)
	lgr, err := app.Test(lg)
	if err != nil {
		t.Fatalf("logout failed: %v", err)
	}
	if lgr.StatusCode != fiber.StatusSeeOther {
		b, _ := io.ReadAll(lgr.Body)
		t.Fatalf("expected 303, got %d (%s)", lgr.StatusCode, string(b))
	}

	// GET /auth/me should be 401
	me2 := httptest.NewRequest("GET", "/auth/me", nil)
	if sessPair != "" {
		me2.Header.Set("Cookie", sessPair)
	}
	mer2, err := app.Test(me2)
	if err != nil {
		t.Fatalf("/auth/me after logout failed: %v", err)
	}
	if mer2.StatusCode != fiber.StatusUnauthorized {
		b, _ := io.ReadAll(mer2.Body)
		t.Fatalf("expected 401, got %d (%s)", mer2.StatusCode, string(b))
	}
}

func TestAuth_Login_InvalidCredentials_SkipIfNoDB(t *testing.T) {
	if !hasDatabase() {
		t.Skip("DATABASE_URL not set; skipping auth integration test")
	}
	app := buildAuthTestApp()

	// Seed CSRF cookie
	g := httptest.NewRequest("GET", "/", nil)
	gr, err := app.Test(g)
	if err != nil {
		t.Fatalf("seed GET failed: %v", err)
	}
	csrfPair := extractCookiePairFromValues(gr.Header.Values("Set-Cookie"), "_gforge_csrf")
	if csrfPair == "" {
		t.Fatalf("missing csrf cookie")
	}
	csrfToken := strings.SplitN(csrfPair, "=", 2)[1]

	// Login with invalid credentials
	form := url.Values{}
	form.Set("email", "nope@example.com")
	form.Set("password", "wrongpass")
	req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-CSRF-Token", csrfToken)
	req.Header.Set("Cookie", csrfPair)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 401, got %d (%s)", resp.StatusCode, string(b))
	}
}
