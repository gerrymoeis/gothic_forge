# Gothic Forge

All batteries included. Minimal friction. Focus on building.

Gothic Forge is a Go web starter built with Go, [Fiber](https://github.com/gofiber/fiber),
[Templ](https://github.com/a-h/templ), [HTMX](https://htmx.org/),
[`gotailwindcss`](https://github.com/gotailwindcss/tailwind) (pure‑Go Tailwind),
and [DaisyUI](https://daisyui.com/). It ships secure defaults, server‑side
rendering, and a fast developer experience with hot reload.

Main stack:

- Go
- Fiber
- Templ
- HTMX
- gotailwindcss (pure‑Go Tailwind)
- DaisyUI

## Features

- **Secure-by-default middleware**: Helmet, CSRF, request ID, rate limiter,
- **SSR with Templ**: Components in `app/templates/` compiled to Go.
- **Pure Go Tailwind CSS**: No Node required. `gotailwindcss` builds
  `app/static/styles.css` from `app/static/tailwind.input.css`.
- **Hot reload**: `gforge dev` runs Templ (watch), `gotailwindcss` (rebuild) and
  [Air](https://github.com/air-verse/air) for server reloads.
- **SEO basics**: SVG favicon, meta tags (Open Graph, Twitter), `/robots.txt`
  and `/sitemap.xml` served at the root. Optional JSON‑LD support via
  `SEO.JSONLD` in `LayoutSEO`; canonical URLs are resolved using `BASE_URL`
  when you pass a path (e.g. `/about`).
- **Clean routing**: Single entry `app/routes/routes.go` plus per‑page
  registrants via `routes.RegisterRoute` for scalable scaffolding.
- **Tests UX**: `gforge test` discovers only packages that have tests to avoid
  noisy `? [no test files]` output.

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
- `/counter/widget` — HTMX widget partial (server‑rendered)
- `POST /counter/increment` — Session‑backed counter increment
- `/favicon.ico` — 301 → `/static/favicon.svg`
- `/robots.txt` — Static robots
- `/sitemap.xml` — Static sitemap
- `/static/*` — Static files (CSS, favicon, etc.)

Main entry for routes: `app/routes/routes.go`.
Additional routes can be registered via per‑page registrants created by the
scaffolder (see “Scaffolding pages”).

## Scaffolding

- `gforge add page <name>`

  Creates a new Templ page and wires it automatically:

  - `app/templates/<name>.templ`
  - `app/routes/<name>_route.go` (registrant via `routes.RegisterRoute`)
  - `app/ssg/register_<name>.go` (included in static export)

  Example:

  ```powershell
  go run ./cmd/gforge add page about
  # Visit http://127.0.0.1:8080/about
  ```

- `gforge add api <name>`

  Creates a minimal JSON API endpoint `GET /api/<name>` with a standard envelope:

  - `app/routes/<kebab-name>_api.go`

  Example:

  ```powershell
  go run ./cmd/gforge add api example
  # Visit http://127.0.0.1:8080/api/example
  # -> {"ok": true, "data": {"message": "Example endpoint"}}
  ```

- `gforge add db`

  Creates a pure-SQL migrations folder:

  - `app/db/migrations/`
  - `app/db/migrations/0001_init.sql` (only if folder was empty)

  Example:

  ```powershell
  go run ./cmd/gforge add db
  ```

Notes:

- Registrants keep `app/routes/routes.go` clean and stable.
- For CI, you can skip Templ generation inside `gforge add page` by setting `GFORGE_SKIP_TEMPL=1`.

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

Tests are centralized under the `tests/` folder for a clean app surface. Use the
CLI to run only packages that contain tests (no `? [no test files]` noise):

```powershell
go run ./cmd/gforge test
```

If you run the raw Go command, it’s normal to see `? [no test files]` lines for
packages without tests:

```powershell
go test ./...
```

## Production build

Build the server binary:

```powershell
go build -o build/server.exe ./cmd/server
```

Run with environment variables (optional):

```powershell
$env:BASE_URL = "https://example.com"
$env:LOG_FORMAT = "json"
$env:CORS_ORIGINS = "https://example.com,https://admin.example.com"
$env:FIBER_HOST = "0.0.0.0"
$env:FIBER_PORT = "8080"
./build/server.exe
```

Prepare CSS for production (pure Go):

```powershell
gotailwindcss build -o app/static/styles.css app/static/tailwind.input.css
```

Update SEO assets for production:

- Edit meta tags in `app/templates/layout_seo.templ` (title, description, OG/Twitter)
- Update absolute URLs in `app/static/robots.txt` and `app/static/sitemap.xml`

By default, Gothic Forge does not inject JSON‑LD or inline scripts to keep security tight and output minimal. If you need structured data, add a small component explicitly where required.

## Environment

Copy `.env.example` to `.env` and customize:

```
APP_ENV=development
FIBER_HOST=127.0.0.1
FIBER_PORT=8080
CSRF_SECRET=change-me
DATABASE_URL=postgres://user:pass@localhost:5432/gforge?sslmode=disable
```

### Configuration

Additional environment variables:

- `BASE_URL` — Used by `LayoutSEO` to resolve canonical links and og:url.
  If `SEO.Canonical` is a path (e.g., `/about`), it will be prefixed with
  this base. If you pass a full URL, it is used as-is. Example: `https://example.com`.
- `LOG_FORMAT` — Set to `json` for JSON logs, otherwise plain text.
- `CORS_ORIGINS` — Comma‑separated list of allowed origins in production (e.g., `https://example.com,https://admin.example.com`). If empty, defaults to permissive (good for dev).

- `RATE_LIMIT_MAX` — Max requests allowed per window. Default: `120`.
- `RATE_LIMIT_WINDOW_SECONDS` — Window duration in seconds. Default: `60`.
- `STATIC_CACHE_SECONDS` — Cache-Control max-age for files under `/static`.
  In development, the default is `no-store` unless you set a value. In
  production, the default is `86400` (1 day) if unset. ETag is also applied to
  `/static` responses.

#### .env usage

Copy `.env.example` to `.env` and restart the server after any change so new
values are picked up:

```powershell
Copy-Item .env.example .env -Force
go run ./cmd/gforge dev
```

#### Verify rate limiting

Set small values locally and restart the dev server:

```env
RATE_LIMIT_MAX=10
RATE_LIMIT_WINDOW_SECONDS=10
```

Use the counter button quickly or run a quick loop:

```powershell
for ($i=1; $i -le 15; $i++) {
  try { (Invoke-WebRequest -UseBasicParsing -Method POST http://127.0.0.1:8080/counter/increment).StatusCode }
  catch { $_.Exception.Response.StatusCode.value__ }
}
```

Expect `429` after the threshold within the same window.

#### Verify static caching

Set a small cache window (e.g., 30s) and restart:

```env
STATIC_CACHE_SECONDS=30
```

Then check headers:

```powershell
$resp = Invoke-WebRequest -Uri http://127.0.0.1:8080/static/app.js -Method GET
$resp.Headers["Cache-Control"]   # public, max-age=30
$resp.Headers["ETag"]            # present
```

## Security

- Strict Content Security Policy (CSP) with no inline scripts by default.
- CSRF protection via header `X-CSRF-Token` and a cookie `_gforge_csrf`. A tiny external script `/static/app.js` sets the header for HTMX requests.
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

## Static Export (SSG)

Gothic Forge can export simple static pages for hosting on any static platform:

```powershell
go run ./cmd/gforge export -o dist
```

This renders selected routes (e.g., `/` and `/counter`) to HTML and copies `app/static/` into `dist/static/`.

## CI/CD & Releases

GitHub Actions workflow at `.github/workflows/ci.yml` runs on Ubuntu, macOS,
and Windows:

- go vet
- go test ./... -v
- govulncheck ./...

Tagged releases are built by `.github/workflows/release.yml` using GoReleaser
with configuration in `.goreleaser.yaml`.

To generate a release:

```
git tag v0.1.0
git push origin v0.1.0
```

Artifacts for `gforge` and the `gothic-forge-server` will be published.

## Deployment examples

See `infra/` for examples:

- `infra/caddy/Caddyfile` — Caddy reverse proxy + static files
- `infra/fly/fly.toml` — Fly.io single-binary deployment example
