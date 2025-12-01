# Contributing to Gothic Forge v3

Thanks for your interest in contributing! This project aims to stay simple and pragmatic.

## Getting started

- Install Go 1.22+
- Optional CLIs: `templ`, `gotailwindcss` (auto-checked by `gforge doctor`)
- Verify environment:

```powershell
go run ./cmd/gforge doctor
```

- Start dev server (Templ watch + Tailwind build + live reload):

```powershell
go run ./cmd/gforge dev
```

- Run tests:

```powershell
go run ./cmd/gforge test --with-build
```

## Branching strategy

- Main branches:
  - `main`: production-ready
  - `develop`: active development
- Feature branches:
  - `feature/<short-desc>`
  - `bugfix/<short-desc>`
  - `hotfix/<short-desc>`
  - `chore/<short-desc>`

## Commit messages

Use Conventional Commits. Keep subject under ~60 chars.

- `feat(scope): add X`
- `fix(scope): correct X`
- `docs(scope): update docs`
- `style(scope): formatting only`
- `refactor(scope): improve structure`
- `test(scope): add/update tests`
- `chore(scope): tooling/build`

## Pull requests

Include:

- What changed and why
- Tests covering changes
- Breaking changes (if any) with migration notes
- Deployment notes (if any)

Checklist:

- [ ] `go run ./cmd/gforge doctor` is green
- [ ] `go run ./cmd/gforge test --with-build` passes
- [ ] Code is formatted (`gofmt`/`goimports`)
- [ ] Docs updated (README/CHANGELOG if applicable)
- [ ] Deployment tested (see deployment testing below)

## Coding guidelines

- Go formatting: `gofmt`/`goimports`
- Keep functions small and focused; add logging for errors
- Prefer explicit error handling; avoid panics
- HTTP handlers return proper status codes and content types
- HTML via Templ components in `app/templates/`
- Tailwind: edit `app/static/tailwind.input.css`; generated CSS lives under `app/styles/`

## Testing

### Unit and Integration Tests

Run the full test suite:

```powershell
go run ./cmd/gforge test --with-build
```

Run specific tests:

```powershell
go test ./tests -v -run TestName
```

Run with coverage:

```powershell
go run ./cmd/gforge test --with-build --coverage
```

### Deployment Testing

Gothic Forge uses an **Opinionated Stack** for production deployments:

- **Compute**: Leapcell
- **Database**: CockroachDB Serverless
- **Cache**: Aiven Valkey
- **CDN/Proxy**: Cloudflare Pages

#### Testing Deployment Changes

Before submitting PRs that affect deployment:

1. **Dry-run validation**:

```powershell
go run ./cmd/gforge deploy --dry-run
```

2. **Test with Leapcell credentials**:

Set up a test environment with the Opinionated Stack:

```powershell
# Required environment variables
LEAPCELL_APP_URL=https://your-test-app.leapcell.dev
COCKROACH_API_KEY=your_cockroach_key
DATABASE_URL=postgresql://...
CLOUDFLARE_API_TOKEN=your_cf_token
CLOUDFLARE_ACCOUNT_ID=your_cf_account
CF_PROJECT_NAME=your-project
JWT_SECRET=your_jwt_secret_min_32_chars
SITE_BASE_URL=https://your-test-app.leapcell.dev
```

3. **Deploy to test environment**:

```powershell
go run ./cmd/gforge deploy
```

4. **Verify health checks**:

```powershell
curl https://your-test-app.leapcell.dev/readyz
```

5. **Test graceful shutdown** (if server changes):

```powershell
# Start server locally
go run ./cmd/server

# In another terminal, send SIGTERM
# Windows: Use Task Manager or taskkill
# Check logs for graceful shutdown messages
```

#### What to Test

- **Deploy command changes**: Run `deploy --dry-run` and verify output
- **Database changes**: Test migrations with CockroachDB
- **Cache changes**: Test with Aiven Valkey connection
- **Server changes**: Test startup, health checks, and graceful shutdown
- **Static assets**: Verify CSS/JS/fonts load correctly through Cloudflare

#### Provider-Specific Notes

Gothic Forge only supports the Opinionated Stack. Do not add support for other providers (Railway, Neon, Back4app, etc.) without discussion in an issue first.

## Security

- Do not open public issues for vulnerabilities
- See `SECURITY.md` for responsible disclosure

## License

By contributing, you agree that your contributions are licensed under the MIT License.
