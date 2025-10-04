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
	paths := []string{}
	if root := findModuleRoot(); root != "" {
		paths = append(paths, filepath.Join(root, ".env"))
	}
	paths = append(paths, ".env")
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			mode := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
			if mode == "" || mode == "development" {
				if err := godotenv.Overload(p); err != nil {
					log.Printf("env: could not overload %s: %v", p, err)
				} else {
					log.Printf("env: loaded %s (overload)", p)
				}
			} else {
				_ = godotenv.Load(p)
			}
			break
		}
	}
	return nil
}

// Get returns env var or default if empty.
func Get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func findModuleRoot() string {
	wd, err := os.Getwd()
	if err != nil { return "" }
	cur := wd
	for {
		if _, err := os.Stat(filepath.Join(cur, "go.mod")); err == nil {
			return cur
		}
		parent := filepath.Dir(cur)
		if parent == cur { return "" }
		cur = parent
	}
}
