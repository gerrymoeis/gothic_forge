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

## Coding guidelines

- Go formatting: `gofmt`/`goimports`
- Keep functions small and focused; add logging for errors
- Prefer explicit error handling; avoid panics
- HTTP handlers return proper status codes and content types
- HTML via Templ components in `app/templates/`
- Tailwind: edit `app/static/tailwind.input.css`; generated CSS lives under `app/styles/`

## Security

- Do not open public issues for vulnerabilities
- See `SECURITY.md` for responsible disclosure

## License

By contributing, you agree that your contributions are licensed under the MIT License.
