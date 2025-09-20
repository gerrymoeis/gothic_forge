package env

import (
  "os"
  "testing"
)

func TestLoadNoEnvFile(t *testing.T) {
  // Ensure Load does not error if .env is missing
  if err := Load(); err != nil {
    t.Fatalf("Load() unexpected error: %v", err)
  }
}

func TestGet(t *testing.T) {
  const key = "GF_TEST_ENV_KEY"
  // Ensure default returned when not set
  if v := Get(key, "default"); v != "default" {
    t.Fatalf("expected default, got %q", v)
  }
  // Ensure value returned when set
  os.Setenv(key, "value")
  defer os.Unsetenv(key)
  if v := Get(key, "default"); v != "value" {
    t.Fatalf("expected value, got %q", v)
  }
}
