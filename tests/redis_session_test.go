package tests

import (
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v2"
	"gothicforge/app/routes"
	"gothicforge/internal/server"
)

func TestSessions_Redis_Backend(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis failed to start: %v", err)
	}
	defer mr.Close()

	// Point session storage to in-memory Redis
	os.Setenv("REDIS_URL", fmt.Sprintf("redis://%s/0", mr.Addr()))
	t.Cleanup(func() { os.Unsetenv("REDIS_URL") })

	app := server.New()
	routes.Register(app)

	// Seed cookies via GET /
	g := httptest.NewRequest("GET", "/", nil)
	gr, err := app.Test(g)
	if err != nil {
		t.Fatalf("seed GET failed: %v", err)
	}
	if gr.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", gr.StatusCode)
	}
	sc := gr.Header.Values("Set-Cookie")
	csrfPair := extractCookiePairFromValues(sc, "_gforge_csrf")
	sessPair := extractCookiePairFromValues(sc, "session")
	if csrfPair == "" {
		t.Fatalf("missing csrf cookie in response")
	}
	token := strings.SplitN(csrfPair, "=", 2)[1]

	// First increment
	req1 := httptest.NewRequest("POST", "/counter/increment", nil)
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
	if resp1.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp1.StatusCode)
	}
	b1, _ := io.ReadAll(resp1.Body)
	s1 := strings.TrimSpace(string(b1))
	if s1 == "" {
		t.Fatalf("empty response")
	}
	// Refresh session cookie if present
	sessValues := resp1.Header.Values("Set-Cookie")
	sessCookie := extractCookiePairFromValues(sessValues, "session")
	if sessCookie != "" {
		semi := strings.Index(sessCookie, ";")
		if semi > 0 {
			sessPair = sessCookie[:semi]
		} else {
			sessPair = sessCookie
		}
	}

	// Second increment
	req2 := httptest.NewRequest("POST", "/counter/increment", nil)
	req2.Header.Set("X-CSRF-Token", token)
	if sessPair != "" {
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
	b2, _ := io.ReadAll(resp2.Body)
	s2 := strings.TrimSpace(string(b2))
	if s2 == "" {
		t.Fatalf("empty response")
	}

	n1, err := strconv.Atoi(s1)
	if err != nil {
		t.Fatalf("invalid number in first response: %q", s1)
	}
	n2, err := strconv.Atoi(s2)
	if err != nil {
		t.Fatalf("invalid number in second response: %q", s2)
	}
	if n2 != n1+1 {
		t.Fatalf("expected second count to be first+1, got %d then %d", n1, n2)
	}
}
