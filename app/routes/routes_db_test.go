package routes

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"gothicforge/internal/server"
)

func TestDBPing_NoConfig(t *testing.T) {
	// Ensure DATABASE_URL is not set
	old := os.Getenv("DATABASE_URL")
	_ = os.Unsetenv("DATABASE_URL")
	defer func() {
		if old != "" {
			os.Setenv("DATABASE_URL", old)
		}
	}()

	app := server.New()
	Register(app)

	req := httptest.NewRequest("GET", "/db/ping", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("/db/ping request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}
