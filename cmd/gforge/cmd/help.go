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

Deploy:
  deploy    Deploy the app (default: Render)
  deploy pages      Deploy static site in dist/ to Cloudflare Pages (via Wrangler)
  deploy provision  Provision the omakase stack (Render; validate provider tokens)

Tools:
  tools list                 List required/optional external tools and their status
  tools install [name|all|deploy]
                            Install a Go-based tool, all, or the deploy stack (Render CLI)

Examples:
  gforge tools install deploy
  gforge deploy provision --app <app> --service <svc>
  gforge deploy --provider render --app <app> --service <svc>

Other:
  version   Print CLI version
  lint      Run basic linting (vet + gofmt)
  doctor    Check environment & dependencies
  release   Build release artifacts with GoReleaser
  add       Scaffold stubs (WIP)

Use "gforge [command] --help" for more information about a command.
`)
}
