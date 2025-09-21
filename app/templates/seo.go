package templates

// SEO holds per-page SEO metadata.
type SEO struct {
	Title       string
	Description string
	Keywords    []string
	Canonical   string
	// JSONLD allows injecting a raw JSON-LD script if needed.
	JSONLD string
}
