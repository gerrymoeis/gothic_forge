package env

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Load loads environment variables from a .env file if present.
// It is safe to call multiple times.
func Load() error {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			log.Printf("warn: could not load .env: %v", err)
		}
	}
	return nil
}

// Get reads an env var or returns a default if not set.
func Get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
