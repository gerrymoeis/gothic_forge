package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func Test_CLI_Export_Generates_Dist(t *testing.T) {
	// Ensure tools are skipped in CI or local runs without tools available
	_ = os.Setenv("GFORGE_SKIP_TOOLS", "1")

	outDir := t.TempDir()
	cmd := exec.Command("go", "run", "./cmd/gforge", "export", "-o", outDir)
	cmd.Dir = ".."
	cmd.Env = append(os.Environ(), "LOG_FORMAT=off")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("export failed: %v", err)
	}

	// index.html for root should exist at dist-test/index.html
	p := filepath.Join(outDir, "index.html")
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("expected %s to exist: %v", p, err)
	}
	// static assets folder should be copied
	if _, err := os.Stat(filepath.Join(outDir, "static")); err != nil {
		t.Fatalf("expected %s to exist: %v", filepath.Join(outDir, "static"), err)
	}
}
