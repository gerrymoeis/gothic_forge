package tests

import (
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"
)

func TestAddAPI_Scaffold_CreatesRegistrant(t *testing.T) {
    root := findModuleRoot(t)
    name := "api_test_sample"
    // The CLI sanitizes names to kebab-case. Mirror that here for filename expectation.
    sanitized := sanitizeKebabForTest(name)

    cmd := exec.Command("go", "run", "./cmd/gforge", "add", "api", name)
    cmd.Dir = root
    out, err := cmd.CombinedOutput()
    if err != nil {
        t.Fatalf("gforge add api failed: %v\n%s", err, string(out))
    }

    routePath := filepath.Join(root, "app", "routes", sanitized+"_api.go")
    t.Cleanup(func() { _ = os.Remove(routePath) })

    b, err := os.ReadFile(routePath)
    if err != nil {
        t.Fatalf("expected API registrant created at %s: %v", routePath, err)
    }
    s := string(b)
    if !strings.Contains(s, "/api/"+sanitized+"\"") || !strings.Contains(s, "RegisterRoute(") {
        t.Fatalf("API registrant content did not include expected bits: %s", routePath)
    }
}

// Minimal mirror of CLI sanitize: lowercase, replace any non [a-z0-9-] with '-', trim '-'
func sanitizeKebabForTest(s string) string {
    s = strings.ToLower(s)
    // Replace underscores and any non-allowed chars with '-'
    b := make([]rune, 0, len(s))
    for _, r := range s {
        if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
            b = append(b, r)
        } else {
            b = append(b, '-')
        }
    }
    out := string(b)
    out = strings.Trim(out, "-")
    // Collapse duplicate '-'
    for strings.Contains(out, "--") {
        out = strings.ReplaceAll(out, "--", "-")
    }
    return out
}
