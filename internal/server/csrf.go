package server

import (
	"net/http"
	"net/url"
	"strings"
)

// CSRFMiddleware provides a simple same-origin check for state-changing requests in production.
// Allowed without checks: GET, HEAD, OPTIONS. For others, require Origin or Referer to match Host.
func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := r.Method
		if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		// Prefer Origin header; fall back to Referer.
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin != "" {
			if !sameOrigin(origin, r.Host) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
			return
		}
		ref := strings.TrimSpace(r.Header.Get("Referer"))
		if ref != "" {
			u, err := url.Parse(ref)
			if err != nil || !sameOrigin(u.Scheme+"://"+u.Host, r.Host) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
			return
		}
		// No Origin/Referer provided; reject to be safe.
		http.Error(w, "forbidden", http.StatusForbidden)
	})
}

func sameOrigin(origin, host string) bool {
	// Accept http(s)://<host> exact host match.
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return strings.EqualFold(u.Host, host)
}
