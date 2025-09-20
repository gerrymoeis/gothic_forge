package security

import (
	"crypto/rand"
	"encoding/base64"
)

// RandomString returns a URL-safe base64 string of n bytes of entropy.
func RandomString(n int) string {
	if n <= 0 {
		n = 16
	}
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
