# Gothic Forge

All batteries included. Minimal friction. Focus on building.

Gothic Forge is a Go web starter that ships secure defaults, server-side
rendering with [Templ](https://github.com/a-h/templ), a pure-Go Tailwind CSS
pipeline via [`gotailwindcss`](https://github.com/gotailwindcss/tailwind), and a
fast developer experience with hot reload.

## Features

- **Secure-by-default middleware**: Helmet, CSRF, request ID, rate limiter,
  compression, CORS, recover.
- **SSR with Templ**: Components in `app/templates/` compiled to Go.
- **Pure Go Tailwind CSS**: No Node required. `gotailwindcss` builds
  `app/static/styles.css` from `app/static/tailwind.input.css`.
- **Hot reload**: `gforge dev` runs Templ (watch), `gotailwindcss` (rebuild) and
  [Air](https://github.com/air-verse/air) for server reloads.
- **SEO basics**: SVG favicon, meta tags (Open Graph, Twitter), `/robots.txt`
  and `/sitemap.xml` served at the root.
- **Clean routing**: Minimal routes under `app/routes/`.

## Quick start

Prerequisites:

- Go 1.22+ (or later)

Recommended CLIs (doctor checks them and dev will auto-install when possible):

- `templ`, `air`, `gotailwindcss`, `govulncheck`, `vegeta`

Run the environment doctor:

```powershell
go run ./cmd/gforge doctor
```

Start development (Templ watch + gotailwindcss build/rebuild + Air):

```powershell
go run ./cmd/gforge dev
```

Then open:

- http://127.0.0.1:8080/
- http://127.0.0.1:8080/healthz
- http://127.0.0.1:8080/robots.txt
- http://127.0.0.1:8080/sitemap.xml

## Tailwind (pure Go)

- Input: `app/static/tailwind.input.css`
- Output (generated): `app/static/styles.css` (git-ignored)

Manual build (optional):

```powershell
gotailwindcss build -o app/static/styles.css app/static/tailwind.input.css
```

## Routes

- `/` — Home (Templ component: `templates.Index()`)
- `/healthz` — JSON health check `{ "ok": true }`
- `/favicon.ico` — 301 → `/static/favicon.svg`
- `/robots.txt` — Static robots, points to `/sitemap.xml`
- `/sitemap.xml` — Static sitemap (update URLs for production)
- `/static/*` — Static files (CSS, favicon, etc.)

## 404 page

A minimal themed 404 is wired as a fallback after routes in
`cmd/server/main.go`. It only renders when the status is 404 so it won’t
intercept valid routes.

## Project layout

```
app/
  routes/         # app routes (Fiber handlers)
  static/         # static files (favicon.svg, tailwind.input.css, styles.css)
  templates/      # Templ components (.templ) and generated .go
cmd/
  gforge/         # CLI (doctor, dev, etc.)
  server/         # main web server entrypoint
internal/
  framework/      # shared server/env/exec helpers
  server/         # alternate server entry (kept for compatibility)
```

## Testing

Basic route tests live in `app/routes/routes_test.go`.

```powershell
go test ./app/routes -v
```

## Production build

Build the server binary:

```powershell
go build -o build/server.exe ./cmd/server
```

Run with environment variables (optional):

```powershell
$env:FIBER_HOST = "0.0.0.0"
$env:FIBER_PORT = "8080"
./build/server.exe
```

Prepare CSS for production (pure Go):

```powershell
gotailwindcss build -o app/static/styles.css app/static/tailwind.input.css
```

Update SEO assets for production:

- Edit meta tags in `app/templates/layout.templ` (title, description, OG/Twitter)
- Update absolute URLs in `app/static/robots.txt` and `app/static/sitemap.xml`

## Security

- Run `govulncheck` periodically.
- Follow standard Go security best practices and keep dependencies updated.
- See `SECURITY.md` for disclosure policy.

## Contributing

Contributions are welcome! Please read `CONTRIBUTING.md` and follow the
Conventional Commits style for commit messages.

## License

MIT — see `LICENSE`.

## Acknowledgements

- [Fiber](https://github.com/gofiber/fiber)
- [Templ](https://github.com/a-h/templ)
- [gotailwindcss](https://github.com/gotailwindcss/tailwind)
