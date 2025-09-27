package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func pathExists(p string) bool {
    _, err := os.Stat(p)
    return err == nil
}

func TestAddDB_Scaffold_MigrationsFolder(t *testing.T) {
	root := findModuleRoot(t)
	migrationsDir := filepath.Join(root, "app", "db", "migrations")
	oldExists := pathExists(migrationsDir)
	// Ensure cleanup only if we created migrations in this test
	t.Cleanup(func() {
		if !oldExists {
			_ = os.RemoveAll(migrationsDir)
		}

	})

	cmd := exec.Command("go", "run", "./cmd/gforge", "add", "db")
	cmd.Dir = root
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gforge add db failed: %v\n%s", err, string(out))
	}

	if _, err := os.Stat(migrationsDir); err != nil {
		t.Fatalf("expected migrations directory at %s: %v", migrationsDir, err)
	}

	initFile := filepath.Join(migrationsDir, "0001_init.sql")
	if _, err := os.Stat(initFile); err != nil {
		t.Fatalf("expected initial migration at %s: %v", initFile, err)
	}
}
