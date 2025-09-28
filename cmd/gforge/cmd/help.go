package cmd

func init() {
	// Customize the default help output to document the workflow and commands clearly.
	rootCmd.SetHelpTemplate(`Gothic Forge CLI — batteries-included developer experience

Usage:
  gforge [command]

Available Commands:
{{range .Commands}}{{if (and (not .Hidden))}}  {{rpad .Use .NamePadding}} {{.Short}}
{{end}}{{end}}

Dev Workflow:
  dev       Start dev server with hot reload (templ watch + Tailwind build)
  test      Run unit tests (auto-generates templ once before running)
  build     Build production server binary (auto-generates templ first)
  export    Export static pages to dist/ (SSG)

Other:
  version   Print CLI version
  lint      Run basic linting (vet + gofmt)
  doctor    Check environment & dependencies
  release   Build release artifacts with GoReleaser
  add       Scaffold stubs (WIP)

Use "gforge [command] --help" for more information about a command.
`)
}
