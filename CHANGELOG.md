# Changelog

## [9.4.0] - 2025-12-03

### 🎯 Major Refactoring: Minimal Framework

This release represents a significant refactoring of Gothic Forge to achieve its original vision: **minimal yet powerful**.

### ✨ What Changed

#### Phase 1: Email Package Simplification
- **Removed**: AWS SES, SendGrid, and Mailgun implementations
- **Kept**: SMTP (universal) and MailHog (development)
- **Result**: 75% code reduction, 19 dependencies removed
- **Migration**: All providers can still be used via SMTP

#### Phase 2: Jobs System Extraction
- **Removed**: Built-in asynq job queue implementation
- **Kept**: Minimal interfaces for flexibility
- **Result**: 99% code reduction, 3 dependencies removed
- **Migration**: Use Asynq, Machinery, or River directly

#### Phase 3: Provider Simplification
- **Removed**: SQLite, CockroachDB-specific code, platform providers (Docker, Leapcell, Cloudflare)
- **Kept**: PostgreSQL, Redis, Valkey
- **Result**: 68% code reduction, 1 dependency removed
- **Migration**: Use PostgreSQL for database, Redis/Valkey for cache

### 📊 Overall Impact

- **Code Reduction**: 53% (15,550 → 7,320 lines)
- **Dependencies Removed**: 23 packages (50+ → ~28)
- **Binary Size**: ~37% smaller
- **Build Time**: ~33% faster

### 🚀 What's Improved

- **Simpler**: Fewer concepts to learn
- **Faster**: Quicker builds and smaller binaries
- **Flexible**: Choose your own tools (job queues, email providers, deployment platforms)
- **Maintainable**: Less code to maintain and debug
- **Focused**: Core web framework features only

### 📝 Breaking Changes

#### Email
- Removed `NewSendGridProvider()`, `NewMailgunProvider()`, `NewSESProvider()`
- **Migration**: Use `NewSMTPProvider()` with your provider's SMTP settings

#### Jobs
- Removed built-in job queue implementation
- **Migration**: Import `github.com/hibiken/asynq` or similar directly

#### Providers
- Removed SQLite, CockroachDB, Docker, Leapcell, Cloudflare providers
- Removed provider CLI commands (`gforge providers *`)
- **Migration**: Use PostgreSQL for database, configure manually

#### Deployment
- Simplified `gforge deploy` to show generic guidance
- Removed platform-specific deployment automation
- **Migration**: Use Docker, PaaS, or traditional deployment methods

### 🔧 What Stayed

All core framework features remain:
- ✅ HTTP server with chi router
- ✅ Templ templates
- ✅ HTMX support
- ✅ Tailwind CSS (pure Go)
- ✅ Security (CSRF, CSP, rate limiting)
- ✅ Sessions (Redis/Valkey)
- ✅ JWT authentication
- ✅ OAuth support
- ✅ Database migrations (goose)
- ✅ Health checks
- ✅ Hot reload development

### 📚 Documentation

- Updated README with minimal approach
- Updated QUICKSTART with simplified deployment
- All examples updated to reflect new structure

### 🙏 Philosophy

Gothic Forge v9.4 embraces minimalism:
- **Interfaces over implementations**: Provide flexibility
- **Standards over abstractions**: Use SMTP, PostgreSQL, Redis
- **Choice over opinions**: Let developers choose their tools
- **Core over features**: Focus on web framework essentials

### 🔄 Migration Guide

See individual package READMEs for detailed migration instructions:
- `internal/email/README.md` - Email migration
- `internal/jobs/README.md` - Jobs migration
- `internal/providers/` - Provider changes

### 🐛 Bug Fixes

- Fixed MIME type handling in static file serving
- Fixed session store configuration
- Improved error messages throughout

### 🔐 Security

- All security features maintained
- No security regressions
- Smaller attack surface due to less code

---

## [9.3.0] - Previous Release

See git history for previous changes.
