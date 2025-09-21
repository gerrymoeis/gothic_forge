package server

import (
	"net/http/httptest"
	"testing"
)

func TestStaticFavicon(t *testing.T) {
	app := New()
	req := httptest.NewRequest("GET", "/static/favicon.svg", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("static request failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 for /static/favicon.svg, got %d", resp.StatusCode)
	}
}
