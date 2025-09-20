// Code generated placeholder until `templ generate` is run.
package templates

import (
	"github.com/a-h/templ"
	templruntime "github.com/a-h/templ/runtime"
)

// NotFound renders a minimal 404 page (placeholder). Templ will overwrite this on `templ generate`.
func NotFound() templ.Component {
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
		if err = templruntime.WriteString(buf, 1, "<!DOCTYPE html><html lang=\"en\" data-theme=\"dark\"><head><meta charset=\"UTF-8\"/><meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"/><title>Not Found · Gothic Forge</title><link rel=\"icon\" href=\"/static/favicon.svg\" type=\"image/svg+xml\"/><link rel=\"stylesheet\" href=\"/static/styles.css\"/></head><body class=\"min-h-screen bg-base-100 text-base-content grid place-items-center\"><div class=\"text-center p-8\"><h1 class=\"text-6xl font-extrabold bg-gradient-to-r from-indigo-500 to-pink-500 bg-clip-text text-transparent\">404</h1><p class=\"mt-4 opacity-80\">Halaman tidak ditemukan.</p><a href=\"/\" class=\"btn btn-primary mt-6\">Kembali ke Beranda</a></div></body></html>"); err != nil {
			return err
		}
		return nil
	})
}
