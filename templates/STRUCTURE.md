# Gothic Forge Templates - Directory Structure

This document shows the complete structure of all available templates.

## Overview

```
templates/
├── README.md                    # Main templates documentation
├── STRUCTURE.md                 # This file
├── minimal/                     # Minimal template
├── blog/                        # Blog template
├── saas/                        # SaaS template
└── api/                         # API template
```

## Minimal Template

```
templates/minimal/
├── README.md                    # Template documentation
├── .env.example                 # Environment variables
├── providers.yaml               # Provider configuration
├── go.mod.template              # Go module template
└── cmd/
    └── server/
        └── main.go              # Minimal HTTP server
```

**Size**: ~5 files  
**Complexity**: Low  
**Best for**: Learning, prototypes, simple APIs

## Blog Template

```
templates/blog/
├── README.md                    # Template documentation
├── .env.example                 # Environment variables
├── providers.yaml               # Provider configuration
├── go.mod.template              # Go module template
└── app/
    └── db/
        └── migrations/
            └── 001_create_posts.sql  # Database schema
```

**Size**: ~6 files  
**Complexity**: Medium  
**Best for**: Blogs, content sites, documentation

## SaaS Template

```
templates/saas/
├── README.md                    # Template documentation
├── .env.example                 # Environment variables
├── providers.yaml               # Provider configuration
├── go.mod.template              # Go module template
└── app/
    └── db/
        └── migrations/
            ├── 001_create_users.sql         # User schema
            └── 002_create_subscriptions.sql # Billing schema
```

**Size**: ~7 files  
**Complexity**: High  
**Best for**: SaaS apps, multi-tenant platforms, subscription services

## API Template

```
templates/api/
├── README.md                    # Template documentation
├── .env.example                 # Environment variables
├── providers.yaml               # Provider configuration
├── go.mod.template              # Go module template
├── app/
│   └── db/
│       └── migrations/
│           ├── 001_create_api_keys.sql      # API keys schema
│           └── 002_create_resources.sql     # Resources schema
└── docs/
    └── openapi.yaml             # OpenAPI specification
```

**Size**: ~8 files  
**Complexity**: Medium  
**Best for**: REST APIs, microservices, backend services

## Template Comparison

| Feature | Minimal | Blog | SaaS | API |
|---------|---------|------|------|-----|
| **Database** | ❌ | ✅ | ✅ | ✅ |
| **Cache** | ❌ | Optional | ✅ | Optional |
| **Authentication** | ❌ | ❌ | ✅ | ✅ |
| **Frontend** | ❌ | ✅ | ✅ | ❌ |
| **Billing** | ❌ | ❌ | ✅ | ❌ |
| **API Docs** | ❌ | ❌ | ❌ | ✅ |
| **OAuth** | ❌ | ❌ | ✅ | ❌ |
| **Migrations** | ❌ | ✅ | ✅ | ✅ |
| **Files** | 5 | 6 | 7 | 8 |
| **Setup Time** | 1 min | 3 min | 5 min | 3 min |

## Common Files

All templates include:

1. **README.md** - Template-specific documentation
2. **.env.example** - Environment variables with sensible defaults
3. **providers.yaml** - Provider configuration (database, cache, compute, CDN)
4. **go.mod.template** - Go module definition with required dependencies

## Template Variables

When using `gforge init --template=<name>`, these variables are substituted:

- `{{.ProjectName}}` - Project name (from directory or --name flag)
- `{{.ModuleName}}` - Go module name (e.g., github.com/user/project)
- `{{.GitRemote}}` - Git remote URL (if available)
- `{{.JWTSecret}}` - Auto-generated 64-character JWT secret
- `{{.SiteBaseURL}}` - Base URL for the site (from Git remote or default)

## Usage

### Initialize with a template

```bash
# Use minimal template (default)
gforge init

# Use specific template
gforge init --template=blog
gforge init --template=saas
gforge init --template=api
```

### Template selection flow

1. User runs `gforge init --template=<name>`
2. CLI validates template exists
3. CLI copies template files to project directory
4. CLI substitutes template variables
5. CLI runs `go mod download`
6. CLI generates Templ templates (if applicable)
7. CLI builds Tailwind CSS (if applicable)
8. Project is ready for development

## Adding New Templates

To add a new template:

1. Create directory under `templates/`
2. Add required files (README.md, .env.example, providers.yaml, go.mod.template)
3. Add any template-specific files (migrations, handlers, etc.)
4. Update `templates/README.md` with template description
5. Update this file (STRUCTURE.md) with template structure
6. Test with `gforge init --template=your-template`

## Template Best Practices

1. **Keep it minimal** - Only include what's necessary
2. **Document everything** - Clear README with examples
3. **Sensible defaults** - Work out of the box
4. **Provider flexibility** - Support multiple providers
5. **Security first** - Auto-generate secrets, secure defaults
6. **Production ready** - Include migrations, error handling
7. **Easy to customize** - Clear structure, well-commented code

## Next Steps

After initializing with a template:

1. Review the generated `README.md` in your project
2. Configure `.env` with your credentials
3. Run `gforge dev` to start development
4. Customize the template for your needs
5. Deploy with `gforge deploy --live`

## Learn More

- [Main Templates Documentation](README.md)
- [Gothic Forge Documentation](../README.md)
- [Deployment Guide](../QUICKSTART.md)
- [Contributing Guide](../CONTRIBUTING.md)
