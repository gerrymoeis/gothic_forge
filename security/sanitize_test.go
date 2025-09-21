package security

import (
	"strings"
	"testing"
)

func TestSanitizeHTML_RemovesScript(t *testing.T) {
	in := `<p>Hello<script>alert(1)</script></p>`
	out := SanitizeHTML(in)
	if strings.Contains(out, "<script") || strings.Contains(out, "alert(1)") {
		t.Fatalf("expected script content to be removed, got: %s", out)
	}
	if !strings.Contains(out, "Hello") {
		t.Fatalf("expected sanitized output to contain content text, got: %s", out)
	}
}
