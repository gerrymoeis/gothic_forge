//go:build examples
// +build examples

package tests

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
)

type apiResp struct {
	OK   bool `json:"ok"`
	Data struct {
		Message string `json:"message"`
	} `json:"data"`
}

func TestAddAPI_Runtime_RespondsOK(t *testing.T) {
	app := buildTestApp()
	req := httptest.NewRequest("GET", "/api/example", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(b))
	}
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		t.Fatalf("missing Content-Type header")
	}
	var ar apiResp
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&ar); err != nil {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("invalid JSON: %v body=%s", err, string(b))
	}
	if !ar.OK {
		t.Fatalf("expected ok=true in response")
	}
	if ar.Data.Message == "" {
		t.Fatalf("expected data.message to be set")
	}
}
