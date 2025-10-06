# Gothic Forge v3

Lean, batteries-included Go starter. Simple, fast, secure, and great DX.

Gothic Forge v3 is built with Go, [chi](https://github.com/go-chi/chi),
[Templ](https://github.com/a-h/templ), [HTMX](https://htmx.org/),
[`gotailwindcss`](https://github.com/gotailwindcss/tailwind) (pure‑Go Tailwind), and
[DaisyUI](https://daisyui.com/). It ships secure defaults (CSP, CSRF, rate limiting),
server‑side rendering, and a fast developer experience with hot reload.

## Stack

- Go
- chi (router + middlewares)
- Templ (type-safe UI)
- HTMX (progressive interactivity)
- gotailwindcss (pure‑Go Tailwind)
- DaisyUI (via CDN)

## Features

- **Secure-by-default middleware**: Request ID, Real IP, Recoverer, CORS, rate limit (`go-chi/httprate`),
  session cookies (`scs`), CSP, and optional CSRF in production.
- **SSR with Templ**: Components in `app/templates/` rendered on the server.
- **Pure Go Tailwind CSS**: No Node required. `gotailwindcss` produces `app/styles/output.css` from
  `app/styles/tailwind.input.css` (or your inputs).
- **Hot reload**: `gforge dev` runs Templ generation/watch, Tailwind build/rebuild, and reloads the server.
- **SEO basics**: Favicon, meta tags (Open Graph, Twitter), `/robots.txt` and `/sitemap.xml`; JSON‑LD via `LayoutSEO`.
- **Clean routing**: `app/routes/routes.go` mounts core routes; per‑page registrars via `RegisterRoute`.
- **Tests UX**: `gforge test` builds the server first and runs the suite, with quiet logs.

## Quick start

Prerequisites:

- Go 1.22+
- Optional CLIs: `templ`, `gotailwindcss` (auto-checked by `gforge doctor`)

Doctor:

```powershell
go run ./cmd/gforge doctor
```

Dev:

```powershell
go run ./cmd/gforge dev
# Open http://127.0.0.1:8080/
```

Test:

```powershell
go run ./cmd/gforge test --with-build
```

Build:

```powershell
go run ./cmd/gforge build
```

## Routes

- `/` — Home (Templ: `templates.Index()`)
- `POST /counter/sync` — Server-side counter sample
- `/favicon.ico` — 301 → `/static/favicon.svg`
- `/robots.txt` — Defaults or stream `app/static/robots.txt`
- `/sitemap.xml` — Defaults or stream `app/static/sitemap.xml`
- `/readyz` — Readiness (Valkey optional; DB added in Phase 2)
- `/static/*` — Files under `app/static`
- `/static/styles/*` — Files under `app/styles`

Main entry: `app/routes/routes.go`.

## Scaffolding

```powershell
go run ./cmd/gforge add page about
# -> app/templates/page_about.go
# -> app/routes/page_about.go

go run ./cmd/gforge add component Card
# -> app/templates/component_card.go

go run ./cmd/gforge add auth
# -> /login, /logout + template

go run ./cmd/gforge add oauth github
# -> /oauth/github/{start,callback}

go run ./cmd/gforge add db appdata
# -> app/db/appdata.sql

go run ./cmd/gforge add module blog
# -> page + db scaffold
```

## Project layout

```
app/
  routes/      # chi routes and registrars
  static/      # static assets (favicon, tailwind inputs, etc.)
  styles/      # generated CSS and overrides (served at /static/styles)
  templates/   # Templ components (pure Go)
cmd/
  gforge/      # CLI (doctor, dev, build, test, add, etc.)
  server/      # main web server entrypoint
internal/
  env/         # env helpers
  execx/       # exec helpers
  server/      # router constructor, middlewares, CSP, static mounting
```

## Environment

Copy `.env.example` to `.env` and set:

```
APP_ENV=development
HTTP_HOST=127.0.0.1
HTTP_PORT=8080
LOG_FORMAT=
CORS_ORIGINS=
SITE_BASE_URL=http://127.0.0.1:8080
```

- `LOG_FORMAT`: `json` for JSON logs, `off|silent|none` to disable request logs.
- `CORS_ORIGINS`: comma-separated origins (use `*` in dev only).
- `SITE_BASE_URL`: absolute base used by SEO helpers and generated sitemap links.

## Security

- CSP is set per environment. In development, inline script/style is allowed for DX.
  In production, inline style is allowed; scripts are restricted to `self` + known CDNs
  (unpkg/jsDelivr) to support HTMX/Alpine and JSON‑LD where needed.
- CSRF middleware is enabled automatically when `APP_ENV=production`.
- Sessions use secure cookie defaults (`HttpOnly`, `SameSite=Lax`, `Secure` in production).

## CI & Releases

- See `.github/workflows/ci.yml` for vet/test/govulncheck on Windows/macOS/Linux.
- See `.github/workflows/release.yml` + `.goreleaser.yaml` to build `gforge` and `gothic-forge-server`
  on tag push (`v*.*.*`).

## Contributing

See `CONTRIBUTING.md`. Follow Conventional Commits.

## License

MIT — see `LICENSE`.
