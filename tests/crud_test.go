package tests

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gothicforge3/app/routes"
	"gothicforge3/internal/auth"
	"gothicforge3/internal/server"
)

// These tests assume you have generated a CRUD named "posts":
//   go run ./cmd/gforge add crud posts
// If the generated files are missing, tests will be skipped.
func ensurePostsCrudPresent(t *testing.T) {
	if _, err := os.Stat(filepath.Join("..", "app", "routes", "crud_posts.go")); os.IsNotExist(err) {
		t.Skip("crud posts routes not found; run 'gforge add crud posts' to enable this test")
	}
	if _, err := os.Stat(filepath.Join("..", "app", "templates", "crud_posts.go")); os.IsNotExist(err) {
		t.Skip("crud posts templates not found; run 'gforge add crud posts' to enable this test")
	}
}

func Test_CRUD_Posts_Routes_Exist(t *testing.T) {
	ensurePostsCrudPresent(t)
	_ = os.Setenv("LOG_FORMAT", "off")

	r := server.New()
	routes.Register(r)

	// List
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/posts", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /posts want 200, got %d", rec.Code)
	}

	// New
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/posts/new", nil))
	if rec2.Code != http.StatusOK {
		t.Fatalf("GET /posts/new want 200, got %d", rec2.Code)
	}
}

func Test_CRUD_Posts_Create_Unauthorized(t *testing.T) {
	ensurePostsCrudPresent(t)
	_ = os.Setenv("LOG_FORMAT", "off")

	r := server.New()
	routes.Register(r)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader("name=hello&description=world"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("POST /posts (no JWT) want 401, got %d", rec.Code)
	}
}

func Test_CRUD_Posts_Create_Authorized(t *testing.T) {
	ensurePostsCrudPresent(t)
	_ = os.Setenv("LOG_FORMAT", "off")
	_ = os.Setenv("JWT_SECRET", "testsecret")
	auth.Init()

	r := server.New()
	routes.Register(r)

	// Issue JWT cookie
	tok, _, err := auth.Issue(30*time.Minute, map[string]any{"sub": "tester"})
	if err != nil { t.Fatalf("issue token: %v", err) }

	name := "post-" + time.Now().Format("20060102150405")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader("name="+name+"&description=desc"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "gf_jwt", Value: tok, Path: "/"})
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("POST /posts (JWT) want 303, got %d", rec.Code)
	}

	// Verify list includes new item
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/posts", nil))
	if rec2.Code != http.StatusOK { t.Fatalf("GET /posts want 200, got %d", rec2.Code) }
	if !strings.Contains(rec2.Body.String(), name) {
		t.Fatalf("expected list to contain %q; body=%q", name, rec2.Body.String())
	}
}
