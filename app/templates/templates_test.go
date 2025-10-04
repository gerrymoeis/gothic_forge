package templates

import (
	"bytes"
	"context"
	"testing"
)

func TestIndex_RendersHTML(t *testing.T) {
	var buf bytes.Buffer
	if err := Index().Render(context.Background(), &buf); err != nil {
		t.Fatalf("render error: %v", err)
	}
	s := buf.String()
	if !contains(s, "Gothic Forge v3") || !contains(s, "Counter Demo") {
		t.Fatalf("missing expected content in output: %q", s[:min(200, len(s))])
	}
}

func contains(s, sub string) bool { return len(s) >= len(sub) && (s == sub || (len(s) > len(sub) && (contains(s[1:], sub) || s[:len(sub)] == sub))) }
func min(a, b int) int { if a < b { return a }; return b }
