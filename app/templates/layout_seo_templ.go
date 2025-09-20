// Code generated placeholder until `templ generate` is run.
package templates

import (
	"github.com/a-h/templ"
	templruntime "github.com/a-h/templ/runtime"
	"strings"
)

// LayoutSEO placeholder to allow building before templ generation.
func LayoutSEO(meta SEO) templ.Component {
	return templruntime.GeneratedTemplate(func(in templruntime.GeneratedComponentInput) (err error) {
		w, ctx := in.Writer, in.Context
		if ctx.Err() != nil {
			return ctx.Err()
		}
		buf, isBuffer := templruntime.GetBuffer(w)
		if !isBuffer {
			defer func() {
				if relErr := templruntime.ReleaseBuffer(buf); err == nil {
					err = relErr
				}
			}()
		}
		// Minimal SEO head (no dynamic canonical to keep placeholder simple)
		if err = templruntime.WriteString(buf, 1, "<!DOCTYPE html><html lang=\"en\" data-theme=\"dark\"><head><meta charset=\"UTF-8\"/><meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"/>"); err != nil { return err }
		if meta.Title != "" { if err = templruntime.WriteString(buf, 1, "<title>"+meta.Title+"</title>"); err != nil { return err } }
		if meta.Description != "" { if err = templruntime.WriteString(buf, 1, "<meta name=\"description\" content=\""+meta.Description+"\"/>"); err != nil { return err } }
		if len(meta.Keywords) > 0 { if err = templruntime.WriteString(buf, 1, "<meta name=\"keywords\" content=\""+strings.Join(meta.Keywords, ", ")+"\"/>"); err != nil { return err } }
		if err = templruntime.WriteString(buf, 1, "<link rel=\"icon\" href=\"/static/favicon.svg\" type=\"image/svg+xml\"/><link rel=\"stylesheet\" href=\"/static/styles.css\"/></head><body class=\"min-h-screen bg-base-100 text-base-content\">"); err != nil { return err }
		// Render children
		children := templ.GetChildren(ctx); if children == nil { children = templ.NopComponent }
		if err = children.Render(ctx, buf); err != nil { return err }
		if err = templruntime.WriteString(buf, 1, "</body></html>"); err != nil { return err }
		return nil
	})
}
