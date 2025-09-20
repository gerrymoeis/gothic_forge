package security

import (
	"github.com/microcosm-cc/bluemonday"
)

// Sanitizer returns a strict HTML sanitizer suitable for user-generated content
// in limited contexts. Adjust policy as needed per app.
func Sanitizer() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	// Example: allow 
	// p.AllowAttrs("class").OnElements("span", "div")
	return p
}

// SanitizeHTML sanitizes a potentially unsafe HTML string.
func SanitizeHTML(input string) string {
	return Sanitizer().Sanitize(input)
}
