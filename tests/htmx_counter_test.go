package tests

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"gothicforge3/app/routes"
	"gothicforge3/internal/server"
)

func Test_HTMX_CounterSync(t *testing.T) {
	_ = os.Setenv("LOG_FORMAT", "off")
	r := server.New()
	routes.Register(r)
	form := url.Values{}
	form.Set("count", "7")
	req := httptest.NewRequest(http.MethodPost, "/counter/sync", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if strings.TrimSpace(rec.Body.String()) != "7" {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}
