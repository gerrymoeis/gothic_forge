package ssg

import (
	"github.com/a-h/templ"
	"gothicforge/app/templates"
)

// Default registrations for SSG so export works out-of-the-box.
func init() {
	Register("/", ToHTMLFunc(func() templ.Component { return templates.Index() }))
}
