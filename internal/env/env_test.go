package env

import (
	"testing"
)

func TestGet_DefaultAndOverride(t *testing.T) {
	t.Setenv("SOME_KEY", "")
	if got := Get("SOME_KEY", "fallback"); got != "fallback" {
		t.Fatalf("want fallback, got %q", got)
	}
	t.Setenv("SOME_KEY", "value")
	if got := Get("SOME_KEY", "fallback"); got != "value" {
		t.Fatalf("want value, got %q", got)
	}
}

func TestLoad_SafeMultipleCalls(t *testing.T) {
	if err := Load(); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if err := Load(); err != nil {
		t.Fatalf("Load returned error on second call: %v", err)
	}
}
