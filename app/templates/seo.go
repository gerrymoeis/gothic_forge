package templates

import (
    "fmt"
    "strings"
    "gothicforge/internal/env"
    "github.com/a-h/templ"
)

// SEO holds per-page SEO metadata.
type SEO struct {
    Title       string
    Description string
    Keywords    []string
    Canonical   string
    // JSONLD allows injecting a raw JSON-LD script if needed.
    JSONLD string
}

// JSONLDRaw wraps a JSON string as a templ component without escaping.
func JSONLDRaw(json string) templ.Component {
    return templ.Raw(json)
}

// JoinKeywords renders a CSV of keywords for meta tags.
func JoinKeywords(keywords []string) string {
    if len(keywords) == 0 {
        return ""
    }
    return strings.Join(keywords, ", ")
}

// WebSiteJSONLD returns a minimal JSON-LD for a WebSite entity, using
// ResolveCanonical to ensure the URL is absolute when BASE_URL is set.
func WebSiteJSONLD(name, path string) string {
    url := ResolveCanonical(path)
    // Build a minimal JSON object as a string; consumers can extend as needed.
    return fmt.Sprintf(`{"@context":"https://schema.org","@type":"WebSite","url":"%s","name":"%s"}`, url, name)
}

// ResolveCanonical turns a path-only canonical (e.g., "/about") into a fully
// qualified URL using BASE_URL, if provided. If the input already looks like an
// absolute URL (http/https), it is returned as-is.
func ResolveCanonical(pathOrURL string) string {
    u := strings.TrimSpace(pathOrURL)
    if u == "" {
        return u
    }
    if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
        return u
    }
    // Ensure leading slash for path values.
    if !strings.HasPrefix(u, "/") {
        u = "/" + u
    }
    base := strings.TrimRight(env.Get("BASE_URL", ""), "/")
    if base == "" {
        // Fall back to path-only if BASE_URL is not configured.
        return u
    }
    return base + u
}
