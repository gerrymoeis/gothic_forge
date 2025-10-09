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
  - `SEO_KEYWORDS` env lets you override the default keywords included by `LayoutSEO`.
  - `sitemap.xml` includes `<lastmod>` for all URLs.
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
- `/readyz` — Readiness (Valkey optional; DB optional if `DATABASE_URL` is set)
- `/db/posts` — Sample DB‑backed feature (requires `DATABASE_URL`; POST/PUT/DELETE require JWT)
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
SEO_KEYWORDS=
DATABASE_URL=
```

- `LOG_FORMAT`: `json` for JSON logs, `off|silent|none` to disable request logs.
- `CORS_ORIGINS`: comma-separated origins (use `*` in dev only).
- `SITE_BASE_URL`: absolute base used by SEO helpers and generated sitemap links.

## Database (Neon) & Migrations

This project uses PostgreSQL via `pgx` and SQL migrations via `goose`.

### 1) Install DB dependencies

```powershell
go get github.com/jackc/pgx/v5@latest
go get github.com/pressly/goose/v3@latest
go mod tidy
```

### 2) Create a Neon project

- Sign up: https://neon.tech
- Create a project and database (e.g., `neondb`).
- Copy the Postgres connection string from the console and set it in `.env`:

```
DATABASE_URL=postgres://<user>:<password>@<host>.neon.tech/<db>?sslmode=require
```

`sslmode=require` is recommended for Neon.

### 3) Create and run migrations

- Create a migration file:

```powershell
go run ./cmd/gforge add migration create_posts
```

- Edit the generated file in `app/db/migrations/` and add SQL, e.g.,

```
-- +goose Up
CREATE TABLE posts (
  id bigserial PRIMARY KEY,
  title text NOT NULL,
  body text NOT NULL,
  created_at timestamptz DEFAULT now(),
  updated_at timestamptz DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS posts;
```

- Apply migrations:

```powershell
go run ./cmd/gforge db --migrate
```

- Check status or reset:

```powershell
go run ./cmd/gforge db --status
go run ./cmd/gforge db --reset
```

### 4) Run dev and verify

```powershell
go run ./cmd/gforge dev
```

- `/readyz` → should show `db: OK` when `DATABASE_URL` is set and reachable.
- `/db/posts` → sample list/form UI backed by Postgres.

Notes:
- Mutations under `/db/posts` require a valid `gf_jwt` cookie (JWT). Use your OAuth flow or wire a dev-only login helper if needed.

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

## Deployment

### First Deploy (quick guide)

1) Prepare `.env`:

```powershell
cp .env.example .env
go run ./cmd/gforge secrets --set SITE_BASE_URL=https://your-domain
go run ./cmd/gforge secrets --set JWT_SECRET=$(openssl rand -hex 32)
```

2) Preflight and fix:

```powershell
go run ./cmd/gforge doctor --fix
```

3) Dry-run to preview steps:

```powershell
go run ./cmd/gforge deploy --dry-run
```

4) Interactive first deploy (provisions Neon/Valkey when tokens present, builds, syncs env to Railway, optional Pages):

```powershell
go run ./cmd/gforge deploy --run
```

If you plan to use GitHub OAuth, see the section below to set it up before prod.

### Token Checklist

- Railway:
  - `RAILWAY_TOKEN` (project token) or `RAILWAY_API_TOKEN` (account/team)
- Neon: `NEON_TOKEN`
- Aiven Valkey: `AIVEN_TOKEN`
- Cloudflare Pages: `CF_API_TOKEN`, `CF_ACCOUNT_ID`, `CF_PROJECT_NAME`
- Optional OAuth: `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `OAUTH_BASE_URL` (defaults to `SITE_BASE_URL`)

Store these in `.env` locally. The deploy wizard can sync the key runtime ones to Railway for you.

### Railway (server/compute)

Use the deploy wizard to guide environment setup and deploy. It checks required secrets and can run an interactive Railway flow.

```powershell
# dry run (no external calls): shows missing secrets and steps
go run ./cmd/gforge deploy --dry-run

# preflight check (no writes/no external actions): validates tools, tokens, env
go run ./cmd/gforge deploy --check

# interactive wizard (first time):
go run ./cmd/gforge deploy --run

# Flags:
#   --init-project   create/link Railway project if missing (wizard will prompt)
#   --project-name   defaults to gothic-forge-v3
#   --service-name   defaults to web
#   --install-tools  attempt to install Railway CLI if missing
```

Required env (typically stored in `.env` or Railway variables):

```
SITE_BASE_URL=https://your-domain
JWT_SECRET=<generated>

# Optional tokens/keys for provider automation
RAILWAY_TOKEN=...          # project token
RAILWAY_API_TOKEN=...      # account/team token (for create/link)
NEON_TOKEN=...
AIVEN_TOKEN=...
CF_API_TOKEN=...
CF_PROJECT_NAME=...
```

### Cloudflare Pages (static)

Export and deploy static HTML to Cloudflare Pages. `_headers` is generated with security/caching defaults.

```powershell
# one-shot deploy with wrangler (if installed)
go run ./cmd/gforge deploy pages --run --project <pages-project-name>

# or dry-run to see the command printed
go run ./cmd/gforge deploy pages --project <pages-project-name>
```

Wrangler install:

Use Homebrew or prebuilt binaries (no Node required):

- macOS: `brew install cloudflare/wrangler/wrangler`
- Windows/Linux: download from https://github.com/cloudflare/wrangler/releases

Notes:
- Export output defaults to `dist/`. Use `--out` to change.
- Security headers (CSP, HSTS, etc.) are written to `dist/_headers`.

### Valkey (Redis-compatible)

Valkey is optional and used for sessions and caching when configured.

Env variables:

```
VALKEY_URL=redis://user:pass@host:port/0
# or REDIS_URL=...
VALKEY_TLS_SKIP_VERIFY=1   # only in dev, if needed
```

`/readyz` will report `valkey: OK|SKIP` automatically.
