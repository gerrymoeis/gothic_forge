package cmd

import (
	"bytes"
	"testing"
)

func TestRootHelp(t *testing.T) {
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute --help: %v", err)
	}
	got := out.String()
	if !contains(got, "Gothic Forge CLI") && !contains(got, "Usage:") {
		t.Fatalf("unexpected help output: %q", got)
	}
}

func contains(s, sub string) bool { return len(s) >= len(sub) && (s == sub || (len(s) > len(sub) && (contains(s[1:], sub) || s[:len(sub)] == sub))) }
