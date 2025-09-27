package ssg

import (
    "github.com/a-h/templ"
    "gothicforge/app/templates"
)

func init() {
    Register("/about", ToHTMLFunc(func() templ.Component { return templates.About() }))
}
