package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func findModuleRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	dir := wd
	for i := 0; i < 6; i++ { // walk up to 6 levels just in case
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		next := filepath.Dir(dir)
		if next == dir {
			break
		}
		dir = next
	}
	t.Fatalf("could not locate go.mod from %s", wd)
	return ""
}

func TestAddPageScaffold_CreatesFiles(t *testing.T) {
	root := findModuleRoot(t)
	name := fmt.Sprintf("scaffold-%d", time.Now().UnixNano())

	cmd := exec.Command("go", "run", "./cmd/gforge", "add", "page", name)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GFORGE_SKIP_TEMPL=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gforge add page failed: %v\n%s", err, string(out))
	}

	tmplPath := filepath.Join(root, "app", "templates", name+".templ")
	sSGPath := filepath.Join(root, "app", "ssg", "register_"+name+".go")
	routePath := filepath.Join(root, "app", "routes", name+"_route.go")

	// Cleanup on exit
	t.Cleanup(func() {
		_ = os.Remove(tmplPath)
		_ = os.Remove(sSGPath)
		_ = os.Remove(routePath)
	})

	if _, err := os.Stat(tmplPath); err != nil {
		t.Fatalf("expected template created at %s: %v", tmplPath, err)
	}
	if _, err := os.Stat(sSGPath); err != nil {
		t.Fatalf("expected SSG registration at %s: %v", sSGPath, err)
	}
	if _, err := os.Stat(routePath); err != nil {
		t.Fatalf("expected route registrant at %s: %v", routePath, err)
	}

	// Light content checks
	if b, err := os.ReadFile(routePath); err == nil {
		s := string(b)
		if !strings.Contains(s, "/"+name+"\"") || !strings.Contains(s, "RegisterRoute(") {
			t.Fatalf("route registrant content did not include expected bits: %s", routePath)
		}
	}
	if b, err := os.ReadFile(sSGPath); err == nil {
		s := string(b)
		if !strings.Contains(s, "Register(\"/"+name+"\"") {
			t.Fatalf("SSG registration content did not include expected Register path: %s", sSGPath)
		}
	}
}
