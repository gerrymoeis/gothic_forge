package ssg

import (
	"bytes"
	"context"
	"sync"

	"github.com/a-h/templ"
	"gothicforge/app/templates"
)

// Renderer renders a page to HTML.
type Renderer func(ctx context.Context) (string, error)

// StaticPage represents a route path and a renderer that returns HTML.
type StaticPage struct {
	Path   string
	Render Renderer
}

var (
	mu    sync.RWMutex
	pages []StaticPage
)

// Register registers a static page for SSG export.
func Register(path string, render Renderer) {
	mu.Lock()
	pages = append(pages, StaticPage{Path: path, Render: render})
	mu.Unlock()
}

// Pages returns a copy of registered pages.
func Pages() []StaticPage {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]StaticPage, len(pages))
	copy(out, pages)
	return out
}

// ToHTMLFunc wraps a function returning a templ.Component into a Renderer.
func ToHTMLFunc(f func() templ.Component) Renderer {
	return func(ctx context.Context) (string, error) {
		var buf bytes.Buffer
		if err := f().Render(ctx, &buf); err != nil {
			return "", err
		}
		return buf.String(), nil
	}
}

// Default registration so export works out-of-the-box.
func init() {
	Register("/", ToHTMLFunc(func() templ.Component { return templates.Index() }))
}
