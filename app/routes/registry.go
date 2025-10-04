package routes

import (
    "fmt"
    "time"

    "github.com/go-chi/chi/v5"
)

var registrars []func(chi.Router)

// RegisterRoute allows features/scaffolds to mount routes without editing core.
func RegisterRoute(fn func(chi.Router)) { registrars = append(registrars, fn) }

func applyRegistrars(r chi.Router) {
	for _, fn := range registrars {
		fn(r)
	}
}

// URL registry for sitemap generation.
type URLInfo struct {
    Path       string
    LastMod    string // RFC3339 or YYYY-MM-DD
    ChangeFreq string // daily|weekly|monthly|...
    Priority   string // 0.0 - 1.0
}

var urlRegistry = map[string]URLInfo{}

// RegisterURL records a path to include in the sitemap.
func RegisterURL(path string) { urlRegistry[path] = URLInfo{Path: path} }

// RegisterURLMeta records a path with metadata.
func RegisterURLMeta(path string, lastMod time.Time, changeFreq string, priority float64) {
    urlRegistry[path] = URLInfo{
        Path:       path,
        LastMod:    lastMod.Format("2006-01-02"),
        ChangeFreq: changeFreq,
        Priority:   fmt.Sprintf("%.1f", priority),
    }
}

// ListURLInfo returns registered URL metadata (includes root by default).
func ListURLInfo() []URLInfo {
    // Ensure root always present
    if _, ok := urlRegistry["/"]; !ok {
        urlRegistry["/"] = URLInfo{Path: "/"}
    }
    out := make([]URLInfo, 0, len(urlRegistry))
    for _, u := range urlRegistry { out = append(out, u) }
    return out
}
