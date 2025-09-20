# Contributing to Gothic Forge

Thanks for your interest in contributing! This project aims to stay simple and pragmatic. Please follow the guidelines below to keep the experience smooth for everyone.

## Getting started

- Install Go 1.22+
- Optional CLIs: `templ`, `air`, `gotailwindcss`, `govulncheck`, `vegeta`
- Verify environment:

```powershell
go run ./cmd/gforge doctor
```

- Start the dev server (Templ watch + gotailwindcss rebuild + Air):

```powershell
go run ./cmd/gforge dev
```

- Run tests:

```powershell
go test ./... -v
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

Include the following in your PR:

- What changed and why
- Tests covering changes
- Breaking changes (if any) with migration notes
- Deployment notes (if any)

PR checklist:

- [ ] `go run ./cmd/gforge doctor` is green
- [ ] `go test ./...` passes
- [ ] Code is formatted (`gofmt`/`goimports`)
- [ ] Docs updated (README/CHANGELOG if applicable)

## Coding guidelines

- Go formatting: use `gofmt`/`goimports`
- Keep functions small and focused; add logging for errors
- Prefer explicit error handling; avoid panics in application code
- HTTP handlers return proper status codes and JSON when applicable
- For HTML, use Templ components in `app/templates/`
- Tailwind: edit `app/static/tailwind.input.css`; `styles.css` is generated

## Security

- Please do not open public issues for vulnerabilities
- See `SECURITY.md` for responsible disclosure process

## License

By contributing, you agree that your contributions are licensed under the MIT License.
