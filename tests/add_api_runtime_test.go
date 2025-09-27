package tests

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

type apiResp struct {
	OK   bool `json:"ok"`
	Data struct {
		Message string `json:"message"`
	} `json:"data"`
}

func TestAddAPI_Runtime_RespondsOK(t *testing.T) {
	root := findModuleRoot(t)
	name := "api_runtime_sample"
	sanitized := sanitizeKebabForTest(name)

	// Ensure the registrant exists (re-use CLI path to generate if absent)
	routePath := filepath.Join(root, "app", "routes", sanitized+"_api.go")
	if !pathExists(routePath) {
		out, err := runCmd(root, "go", "run", "./cmd/gforge", "add", "api", name)
		if err != nil {
			t.Fatalf("gforge add api failed: %v\n%s", err, string(out))
		}
	}
	// Clean up registrant after test
	t.Cleanup(func() { _ = removeFile(routePath) })

	app := buildTestApp()
	req := httptest.NewRequest("GET", "/api/"+sanitized, nil)
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

// small shell helpers for tests
func runCmd(dir string, name string, args ...string) ([]byte, error) {
	cmd := execCommand(name, args...)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

// indirection for exec.Command to allow stubbing in the future
var execCommand = defaultExecCommand

func defaultExecCommand(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

func removeFile(p string) error { return os.Remove(p) }
