package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestIsRailwayLinkedCLI_WithToken(t *testing.T) {
    tmp := t.TempDir() // register TempDir cleanup first
    cwd, _ := os.Getwd()
    t.Cleanup(func() { _ = os.Chdir(cwd) }) // ensure cwd restored before TempDir RemoveAll
    if err := os.Chdir(tmp); err != nil { t.Fatal(err) }
	// Set token to short-circuit
	os.Setenv("RAILWAY_TOKEN", "x")
	t.Cleanup(func() { os.Unsetenv("RAILWAY_TOKEN") })
	if !isRailwayLinkedCLI(context.Background()) {
		t.Fatalf("expected linked with RAILWAY_TOKEN set")
	}
}

func TestIsRailwayLinkedCLI_WithDotRailwayDir(t *testing.T) {
    tmp := t.TempDir() // register TempDir cleanup first
    cwd, _ := os.Getwd()
    t.Cleanup(func() { _ = os.Chdir(cwd) })
    if err := os.Chdir(tmp); err != nil { t.Fatal(err) }
    if err := os.MkdirAll(filepath.Join(tmp, ".railway"), 0o755); err != nil { t.Fatal(err) }
	os.Unsetenv("RAILWAY_TOKEN")
	os.Unsetenv("RAILWAY_API_TOKEN")
	if !isRailwayLinkedCLI(context.Background()) {
		t.Fatalf("expected linked when .railway/ exists")
	}
}
