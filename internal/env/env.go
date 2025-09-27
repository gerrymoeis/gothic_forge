package env

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// Load loads environment variables from a .env file if present.
// It is safe to call multiple times.
func Load() error {
	// Try to find module root (dir containing go.mod) and prefer .env from there.
	// This avoids issues when working directory differs (e.g., via Air).
	envPaths := []string{}
	if root := findModuleRoot(); root != "" {
		envPaths = append(envPaths, filepath.Join(root, ".env"))
	}
	envPaths = append(envPaths, ".env")

	for _, p := range envPaths {
		if _, err := os.Stat(p); err == nil {
			mode := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
			if mode == "" || mode == "development" {
				if err := godotenv.Overload(p); err != nil {
					log.Printf("warn: could not overload %s: %v", p, err)
				} else {
					log.Printf("env: loaded %s (overload)", p)
				}
			} else {
				if err := godotenv.Load(p); err != nil {
					log.Printf("warn: could not load %s: %v", p, err)
				} else {
					log.Printf("env: loaded %s", p)
				}
			}
			// stop at first present .env
			break
		}
	}
	return nil
}

// findModuleRoot walks up from current working directory to locate a folder
// containing go.mod and returns its absolute path, or empty string if not found.
func findModuleRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	cur := wd
	for {
		if _, err := os.Stat(filepath.Join(cur, "go.mod")); err == nil {
			return cur
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return ""
		}
		cur = parent
	}
}

// Get reads an env var or returns a default if not set.
func Get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
