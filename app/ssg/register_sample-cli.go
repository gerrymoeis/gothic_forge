package ssg

import (
	"github.com/a-h/templ"
	"gothicforge/app/templates"
)

func init() {
	Register("/sample-cli", ToHTMLFunc(func() templ.Component { return templates.SampleCli() }))
}
