# Gothic Forge v3

Lean, batteries-included Go starter. Simple, fast, secure, and great DX.

Gothic Forge v3 is built with Go, [chi](https://github.com/go-chi/chi),
[Templ](https://github.com/a-h/templ), [HTMX](https://htmx.org/),
[`gotailwindcss`](https://github.com/gotailwindcss/tailwind) (pure‑Go Tailwind), and
[DaisyUI](https://daisyui.com/). It ships secure defaults (CSP, CSRF, rate limiting),
server‑side rendering, and a fast developer experience with hot reload.

## Philosophy: Opinionated Stack

Gothic Forge v3 embraces an **opinionated, production-ready approach** to web development with a carefully curated stack:

- **Batteries-included** - Sane defaults that work out of the box
- **Production-proven** - Stack choices based on real-world experience at scale
- **Developer-friendly** - Simple deployment with powerful capabilities
- **Cost-effective** - Serverless-first architecture that scales to zero

### The Opinionated Stack

Gothic Forge standardizes on a modern, serverless-first infrastructure:

- **Compute**: [Leapcell](https://leapcell.io/) - Simple, fast container deployment
- **Database**: [CockroachDB Serverless](https://cockroachlabs.cloud/) - Distributed PostgreSQL
- **Cache**: [Aiven Valkey](https://aiven.io/) - Redis-compatible, managed
- **CDN/Proxy**: [Cloudflare](https://cloudflare.com/) - Global edge network

**Why these choices?**
1. **Serverless-first** - Pay only for what you use, scale automatically
2. **PostgreSQL-compatible** - Use familiar tools and patterns
3. **Global by default** - Low latency worldwide
4. **Simple deployment** - One command to production
5. **Production-grade** - Built-in resilience and monitoring

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

- **Go 1.22+** (required)
- **Git** (required for deployments)
- Optional CLIs: `templ`, `gotailwindcss` (auto-checked by `gforge doctor`)

Run `gforge doctor --fix` to check all prerequisites and get installation guidance.

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
- `/db/posts` — Sample DB‑backed feature (requires `DATABASE_URL`; POST/PUT/DELETE require JWT)
- `/static/*` — Files under `app/static`
- `/static/styles/*` — Files under `app/styles`

### Health Check Endpoints

Production-grade health monitoring endpoints following Kubernetes best practices:

- **`/healthz`** — Basic health check (always returns `ok` if app is running)
  - Used by: Docker HEALTHCHECK, uptime monitors
  - Returns: 200 OK with `ok` response

- **`/livez`** — Liveness probe (process health check)
  - Used by: Kubernetes liveness probes
  - Returns: 200 OK with `alive` response
  - Purpose: Container should be restarted if this fails

- **`/readyz`** — Readiness probe (dependency health checks)
  - Used by: Kubernetes readiness probes, load balancers
  - Checks: Database connectivity (if configured), Valkey/Redis (if configured)
  - Returns: 200 OK with detailed status when ready
  - Returns: 503 Service Unavailable when dependencies are down
  - Purpose: Remove pod from load balancer rotation if not ready

Example readiness response:
```
valkey: OK
db: OK
ready
```

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

## Database & Migrations

Gothic Forge uses **CockroachDB Serverless** as the opinionated database standard, with PostgreSQL compatibility via `pgx` and SQL migrations via `goose`.

### Why CockroachDB Serverless?

- **PostgreSQL-compatible** - Works with existing PostgreSQL tools and libraries
- **True serverless** - Pay only for what you use, scales to zero
- **Global distribution** - Low latency worldwide with automatic replication
- **Built-in resilience** - Automatic failover and high availability
- **Production-proven** - Powers mission-critical applications at scale

### 1) Automatic Provisioning (Recommended)

The `gforge deploy` command automatically provisions and configures your database:

```bash
# Set your CockroachDB API key in .env
COCKROACH_API_KEY=your_api_key_here

# Deploy will automatically:
# 1. Create a serverless cluster
# 2. Configure database and user
# 3. Generate secure connection string
# 4. Run migrations automatically
gforge deploy
```

**Get your API key**: https://cockroachlabs.cloud/signup

### 2) Manual Setup (Alternative)

If you prefer manual setup or want to use an existing cluster:

1. Create a CockroachDB Serverless cluster at https://cockroachlabs.cloud
2. Copy the connection string and set it in `.env`:

```bash
DATABASE_URL=postgresql://<user>:<password>@<host>:26257/<db>?sslmode=verify-full
```

**Note**: CockroachDB uses `sslmode=verify-full` for enhanced security.

### 3) Working with Migrations

Migrations are located in `app/db/migrations/` and use the goose format.

#### Create and run migrations

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
- `/db/posts` → sample list/form UI backed by CockroachDB.

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

### The Opinionated Stack

Gothic Forge v3 uses a carefully curated, production-ready stack:

| Component | Provider | Why This Choice |
|-----------|----------|-----------------|
| **Compute** | [Leapcell](https://leapcell.io/) | Simple container deployment, automatic scaling, GitHub integration |
| **Database** | [CockroachDB Serverless](https://cockroachlabs.cloud/) | Distributed PostgreSQL, true serverless, global replication |
| **Cache** | [Aiven Valkey](https://aiven.io/) | Redis-compatible, fully managed, high availability |
| **CDN/Proxy** | [Cloudflare](https://cloudflare.com/) | Global edge network, DDoS protection, automatic SSL |

This stack provides:
- **Serverless-first architecture** - Pay only for what you use
- **Global distribution** - Low latency worldwide
- **Production-grade reliability** - Built-in resilience and monitoring
- **Simple deployment** - One command to production
- **Cost-effective** - Generous free tiers, scales with your needs

### First Deploy (quick guide)

1) Prepare `.env`:

```powershell
cp .env.example .env
go run ./cmd/gforge secrets --set SITE_BASE_URL=https://your-domain
go run ./cmd/gforge secrets --set JWT_SECRET=$(openssl rand -hex 32)
```

2) Get your API keys:

- **CockroachDB**: https://cockroachlabs.cloud/signup
- **Aiven Valkey**: https://console.aiven.io/signup
- **Cloudflare**: https://dash.cloudflare.com/profile/api-tokens
- **Leapcell**: https://leapcell.io/

3) Preflight and fix:

```powershell
go run ./cmd/gforge doctor --fix
```

4) Deploy to production:

```powershell
# Interactive deployment wizard
go run ./cmd/gforge deploy --run

# Or dry-run to see what would happen
go run ./cmd/gforge deploy --dry-run
```

### Environment Variables Checklist

**Required for Deployment**:
- `COCKROACH_API_KEY` - CockroachDB API key for database provisioning
- `AIVEN_TOKEN` - Aiven API token for Valkey cache provisioning
- `CLOUDFLARE_API_TOKEN` - Cloudflare API token for CDN/proxy
- `CLOUDFLARE_ACCOUNT_ID` - Your Cloudflare account ID
- `LEAPCELL_APP_URL` - Your Leapcell application URL (set after first deploy)
- `SITE_BASE_URL` - Your production domain (e.g., https://your-app.com)
- `JWT_SECRET` - Secure random string (min 32 chars)

**Optional**:
- `CF_PROJECT_NAME` - Cloudflare Pages project name (for static exports)
- `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET` - For OAuth integration
- `OAUTH_BASE_URL` - OAuth callback base (defaults to `SITE_BASE_URL`)

Store these in `.env` locally. The deploy wizard will guide you through setting them up.

### Leapcell Deployment

Leapcell provides simple, fast container deployment with automatic scaling and GitHub integration.

**First-time setup**:

```powershell
# Interactive deployment wizard
go run ./cmd/gforge deploy --run
```

The wizard will:
1. Validate your environment configuration
2. Provision CockroachDB database (if needed)
3. Provision Aiven Valkey cache (if needed)
4. Configure Cloudflare proxy settings
5. Deploy your application to Leapcell
6. Run database migrations automatically

**Subsequent deployments**:

```bash
# Quick redeploy after changes
git commit -am "your changes"
go run ./cmd/gforge deploy --run
```

**Deployment options**:

```powershell
# Dry run (see what would happen, no external calls)
go run ./cmd/gforge deploy --dry-run

# Preflight check (validate tools, tokens, env)
go run ./cmd/gforge deploy --check
```

**What gets deployed**:
- Your Go application as a container
- Database migrations (automatic)
- Environment variables (synced securely)
- Health check endpoints configured
- Cloudflare proxy for global CDN

**Troubleshooting**:
- Check `gforge doctor` for prerequisites
- Verify all API keys are set in `.env`
- Check deployment logs in Leapcell dashboard
- Verify `/readyz` endpoint shows all services healthy

### Cloudflare Configuration

Cloudflare provides the CDN/proxy layer for your application, offering:
- Global edge network for low latency
- Automatic DDoS protection
- Free SSL/TLS certificates
- Caching and performance optimization

**Setup**:

1. Get your Cloudflare API token: https://dash.cloudflare.com/profile/api-tokens
2. Set in `.env`:

```bash
CLOUDFLARE_API_TOKEN=your_token_here
CLOUDFLARE_ACCOUNT_ID=your_account_id
```

3. The deploy wizard will configure Cloudflare automatically

**Optional: Static Site Export**

You can also export and deploy static HTML to Cloudflare Pages:

```powershell
# Deploy static export to Cloudflare Pages
go run ./cmd/gforge deploy pages --run --project <pages-project-name>

# Dry-run to see the command
go run ./cmd/gforge deploy pages --project <pages-project-name>
```

Notes:
- Export output defaults to `dist/`. Use `--out` to change.
- Security headers (CSP, HSTS, etc.) are written to `dist/_headers`.

### Aiven Valkey (Redis-compatible Cache)

Aiven Valkey provides a fully managed, Redis-compatible cache for sessions and caching.

**Why Aiven Valkey?**
- Fully managed, no maintenance required
- High availability with automatic failover
- Redis-compatible, works with existing tools
- Generous free tier for development

**Setup**:

1. Get your Aiven API token: https://console.aiven.io/signup
2. Set in `.env`:

```bash
AIVEN_TOKEN=your_token_here
```

3. The deploy wizard will provision Valkey automatically and set `VALKEY_URL`

**Manual configuration** (if using existing Valkey/Redis):

```bash
VALKEY_URL=redis://user:pass@host:port/0
VALKEY_TLS_SKIP_VERIFY=1   # only in dev, if needed
```

The `/readyz` endpoint will report `valkey: OK|SKIP` automatically.
