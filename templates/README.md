# Gothic Forge Project Templates

This directory contains starter templates for `gforge init --template=<name>`.

## Available Templates

### 1. **minimal** (Default)
A bare-bones Gothic Forge project with just the essentials.

**Includes:**
- Basic HTTP server setup
- Health check endpoint
- Minimal routing
- Environment configuration
- No database, no auth, no extras

**Best for:**
- Learning Gothic Forge
- Simple APIs
- Prototypes
- Static sites with minimal backend

**Usage:**
```bash
gforge init --template=minimal
# or just: gforge init (minimal is default)
```

---

### 2. **blog**
A complete blog application with CRUD operations.

**Includes:**
- Post management (Create, Read, Update, Delete)
- Database schema and migrations
- Templ templates for blog UI
- Admin interface
- RSS feed generation
- SEO optimization

**Best for:**
- Personal blogs
- Content sites
- Documentation sites
- News portals

**Usage:**
```bash
gforge init --template=blog
```

---

### 3. **saas**
A full-featured SaaS starter with authentication and billing.

**Includes:**
- User authentication (email/password + OAuth)
- User dashboard
- Subscription management
- Billing integration (Stripe ready)
- Role-based access control (RBAC)
- Email verification
- Password reset flow
- Session management
- Admin panel

**Best for:**
- SaaS applications
- Multi-tenant apps
- Subscription services
- B2B platforms

**Usage:**
```bash
gforge init --template=saas
```

---

### 4. **api**
A REST API-only template with no frontend.

**Includes:**
- RESTful routing structure
- JSON request/response handling
- API versioning (/v1/, /v2/)
- Request validation
- API key authentication
- Rate limiting
- OpenAPI/Swagger documentation
- CORS configuration

**Best for:**
- Backend APIs
- Microservices
- Mobile app backends
- Third-party integrations

**Usage:**
```bash
gforge init --template=api
```

---

## Template Structure

Each template follows this structure:

```
templates/<template-name>/
├── README.md              # Template-specific documentation
├── app/
│   ├── routes/           # HTTP route handlers
│   ├── templates/        # Templ templates (if applicable)
│   ├── static/           # Static assets
│   └── db/
│       └── migrations/   # Database migrations (if applicable)
├── cmd/
│   └── server/
│       └── main.go       # Application entry point
├── internal/             # Internal packages
├── .env.example          # Environment variables template
├── providers.yaml        # Provider configuration
└── go.mod                # Go module definition
```

---

## Creating Custom Templates

To create your own template:

1. Create a new directory under `templates/`
2. Add all necessary files and structure
3. Include a `README.md` explaining the template
4. Update this file with your template description
5. Test with: `gforge init --template=your-template`

---

## Template Variables

Templates support variable substitution during `gforge init`:

- `{{.ProjectName}}` - Project name (from directory or flag)
- `{{.ModuleName}}` - Go module name
- `{{.GitRemote}}` - Git remote URL (if available)
- `{{.JWTSecret}}` - Auto-generated JWT secret
- `{{.SiteBaseURL}}` - Base URL for the site

---

## Provider Configuration

All templates include a `providers.yaml` with sensible defaults:

```yaml
database:
  provider: cockroachdb  # or postgresql, sqlite
  
cache:
  provider: valkey       # or redis, none
  
compute:
  provider: leapcell     # or docker
  
cdn:
  provider: cloudflare   # or none
```

Templates can override these defaults based on their needs:
- **minimal**: No database, no cache
- **blog**: Database required, cache optional
- **saas**: Database + cache required
- **api**: Database required, cache recommended

---

## Next Steps

After initializing with a template:

1. Review the generated `README.md` in your project
2. Configure `.env` with your credentials
3. Run `gforge dev` to start development
4. Run `gforge deploy --live` when ready for production

---

## Contributing

To contribute a new template:

1. Fork the repository
2. Create your template under `templates/`
3. Test thoroughly
4. Submit a pull request
5. Include documentation and examples

See `CONTRIBUTING.md` for more details.
