# 🚀 Gothic Forge v3 - Quickstart

**Development → Production in 2 Minutes**

---

## 📋 Prerequisites

- **Go 1.21+** - [Download](https://go.dev/dl/)
- **Git** - [Download](https://git-scm.com/downloads)

---

## ⚡ Quick Start (3 Steps)

### **1. Clone & Install**

```bash
git clone https://github.com/gerrymoeis/gothic_forge.git
cd gothic_forge

# Build CLI
go build -o gforge ./cmd/gforge  # or gforge.exe on Windows

# Bootstrap project
./gforge install
```

**What this does**:
- ✅ Installs Go tools (templ, gotailwindcss, air)
- ✅ Creates .env with secure JWT_SECRET
- ✅ Sets up Tailwind + static assets
- ✅ Creates Dockerfile

---

### **2. Start Development**

```bash
# Start dev server with hot reload
./gforge dev
```

Visit: `http://localhost:8080`

**Commands**:
```bash
./gforge test          # Run tests
./gforge build         # Build production binary
./gforge doctor        # Check system health
```

---

### **3. Deploy to Production**

#### **Option A: Static Site Only** (Fastest - 1 minute)

Perfect for: Landing pages, blogs, documentation

```bash
# Export static site
./gforge export

# Deploy to Cloudflare Pages
./gforge deploy pages --project=my-site --run
```

**What you get**:
- ✅ Global CDN (300+ locations)
- ✅ HTMX interactions via Pages Functions
- ✅ Sub-50ms responses
- ✅ 100% free tier

**Setup**: Only needs [Cloudflare API token](https://dash.cloudflare.com/profile/api-tokens)

---

#### **Option B: Full Stack (Opinionated Stack)** (Production - ~5 minutes)

Perfect for: Production apps, SaaS, complex applications

Gothic Forge uses an **Opinionated Stack** for production deployments:
- **Compute**: Leapcell (Docker containers)
- **Database**: CockroachDB Serverless (PostgreSQL-compatible)
- **Cache**: Aiven Valkey (Redis-compatible)
- **CDN**: Cloudflare Proxy (orange-cloud in front of Leapcell)

```bash
# One-time: Set up API keys
./gforge secrets --set COCKROACH_API_KEY=<your-key>
./gforge secrets --set AIVEN_TOKEN=<your-token>
./gforge secrets --set CLOUDFLARE_API_TOKEN=<your-token>
./gforge secrets --set LEAPCELL_APP_URL=<your-app-url>

# Deploy (guided)
./gforge deploy --with-valkey
```

**What you get**:
- ✅ Go backend (Leapcell compute)
- ✅ PostgreSQL-compatible database (CockroachDB Serverless)
- ✅ Redis-compatible cache (Aiven Valkey)
- ✅ Cloudflare Proxy CDN (orange‑cloud) in front of Leapcell
- ✅ Global distribution on generous free tiers

---

## 🔑 Get API Keys

### **Leapcell** (Compute) ⭐

1. Sign up: [leapcell.io](https://leapcell.io)
2. Create a new app
3. Copy your app URL (e.g., `https://your-app.leapcell.dev`)
4. Save: `./gforge secrets --set LEAPCELL_APP_URL=<url>`

**Free tier**: Generous compute hours for side projects

### **CockroachDB** (Database) ⭐

1. Sign up: [cockroachlabs.cloud/signup](https://cockroachlabs.cloud/signup)
2. Create service account: [Service Accounts](https://cockroachlabs.cloud/service-accounts)
3. Copy API key (shown once!)
4. Save: `./gforge secrets --set COCKROACH_API_KEY=<key>`

**Free tier**: 5 GB storage, 50M RUs/month

### **Aiven** (Valkey Cache) ⭐

1. Sign up: [console.aiven.io/signup](https://console.aiven.io/signup)
2. Generate token: [Profile → Tokens](https://console.aiven.io/profile/tokens)
3. Copy token (shown once!)
4. Save: `./gforge secrets --set AIVEN_TOKEN=<token>`

**Free trial**: 30 days, then $10/month for Startup plan

### **Cloudflare** (CDN & Static Assets) ⭐

1. Sign up: [dash.cloudflare.com/sign-up](https://dash.cloudflare.com/sign-up)
2. Create API token: [Profile → API Tokens](https://dash.cloudflare.com/profile/api-tokens)
   - Template: "Edit Cloudflare Workers"
   - Permissions: Pages Edit + Workers Edit
3. Copy Account ID: [Dashboard → Account ID](https://dash.cloudflare.com/)
4. Install CLI: `npm install -g wrangler`
5. Save credentials:
   ```bash
   ./gforge secrets --set CLOUDFLARE_API_TOKEN=<token>
   ./gforge secrets --set CLOUDFLARE_ACCOUNT_ID=<account-id>
   ```

**Free tier**: Unlimited static requests, 100k Workers requests/day

---

## 📚 Deployment Paths

### **Path 1: Cloudflare Pages Only**

Best for: Static sites with light interactivity

```bash
./gforge export
./gforge deploy pages --project=my-project --run
```

**Deploy time**: ~1 minute  
**Cost**: $0/month  
**Includes**: Cloudflare Pages Functions for dynamic endpoints

### **Path 2: Full Stack (Opinionated Stack)**

Best for: Production apps, SaaS, complex applications

```bash
./gforge deploy --with-valkey
```

**Deploy time**: ~5 minutes  
**Stack**: Leapcell + CockroachDB + Aiven Valkey + Cloudflare  
**Requires**: API keys from all providers (see above)

---

## 🎯 Which Path is Right for You?

| Need | Choose | Why |
|------|--------|-----|
| **Landing page, blog, docs** | Pages Only | Fastest, free, no backend needed |
| **SaaS, user auth, database** | Full Stack | Production-ready, all features |
| **Global performance** | Full Stack | Cloudflare CDN + edge caching |
| **Scalable architecture** | Full Stack | Serverless database + compute |

---

## 🛠️ CLI Commands

```bash
# Development
./gforge dev                    # Start dev server (hot reload)
./gforge test                   # Run tests
./gforge test --coverage        # With coverage
./gforge build                  # Build production binary
./gforge doctor                 # Check system health

# Deployment
./gforge deploy --check         # Validate secrets/config
./gforge deploy --dry-run       # Preview without executing
./gforge deploy pages --run     # Deploy static site
./gforge deploy --with-valkey   # Full stack deployment

# Database
./gforge db --migrate           # Run migrations
./gforge db --status            # Check migration status

# Secrets
./gforge secrets --gen-jwt      # Generate JWT secret
./gforge secrets --set KEY=val  # Set secret in .env
```

---

## 📦 What Gets Deployed

### **Static Site (Pages Only)**
```
Cloudflare Pages
├── HTML/CSS/JS → CDN (global)
└── functions/
    └── counter/sync.js → Edge endpoint
```

### **Full Stack (Opinionated Stack)**
```
Cloudflare Proxy (CDN + DDoS Protection)
    ↓
Leapcell (Go Backend - Docker Containers)
    ↓
Aiven Valkey (Redis-compatible Cache)
    ↓
CockroachDB Serverless (PostgreSQL Database)
```

---

## 🔄 CI/CD with GitHub Actions

Gothic Forge includes automated deployment workflows:

**Auto-deploy on push to `main`**:
```yaml
# .github/workflows/ci.yml already configured!
# Just add secrets to GitHub repo:
Settings → Secrets → New repository secret

Required secrets:
- CLOUDFLARE_API_TOKEN
- CLOUDFLARE_ACCOUNT_ID
- CF_PROJECT_NAME
- LEAPCELL_APP_URL
- COCKROACH_API_KEY
- AIVEN_TOKEN (optional)
- DATABASE_URL (optional - if not auto-provisioning)
```

**Manual deploy trigger**:
Go to Actions → Manual Deploy → Run workflow

---

## 🚨 Troubleshooting

### Leapcell Deployment Issues

#### **App URL not set**
```bash
# Get your app URL from Leapcell dashboard
./gforge secrets --set LEAPCELL_APP_URL=https://your-app.leapcell.dev
```

#### **Build fails on Leapcell**
```bash
# Verify Dockerfile builds locally
docker build -t test .

# Check logs in Leapcell dashboard
# Common issues: missing dependencies, incorrect PORT binding
```

#### **Database connection fails**
```bash
# Verify DATABASE_URL is set correctly
./gforge doctor

# Check CockroachDB cluster is active in dashboard
# Ensure IP allowlist includes 0.0.0.0/0 for serverless
```

### CockroachDB Issues

#### **API key invalid**
```bash
# Regenerate service account key
# Go to: https://cockroachlabs.cloud/service-accounts
./gforge secrets --set COCKROACH_API_KEY=<new-key>
```

#### **Connection timeout**
```bash
# Check cluster status in CockroachDB dashboard
# Verify DATABASE_URL format:
# postgresql://user:password@host:26257/defaultdb?sslmode=verify-full
```

### Aiven Valkey Issues

#### **Token expired**
```bash
# Generate new token at: https://console.aiven.io/profile/tokens
./gforge secrets --set AIVEN_TOKEN=<new-token>
```

#### **Cache connection fails**
```bash
# Verify VALKEY_URL format:
# redis://default:password@host:port
# Or rediss:// for TLS connections

# Check service status in Aiven dashboard
```

### Cloudflare Issues

#### **API token permissions**
```bash
# Token needs these permissions:
# - Account.Cloudflare Pages: Edit
# - Account.Cloudflare Workers Scripts: Edit

# Regenerate at: https://dash.cloudflare.com/profile/api-tokens
./gforge secrets --set CLOUDFLARE_API_TOKEN=<new-token>
```

#### **Pages deployment fails**
```bash
# Verify wrangler is installed
npm install -g wrangler

# Check account ID is correct
./gforge secrets --set CLOUDFLARE_ACCOUNT_ID=<account-id>
```

### General Issues

#### **Missing tools**
```bash
# Re-run install
./gforge install

# Or use doctor with --fix
./gforge doctor --fix
```

#### **Environment variables not loading**
```bash
# Check .env file exists and has correct format
cat .env  # Linux/Mac
type .env # Windows

# Regenerate if needed
./gforge secrets --gen-jwt
```

#### **Deploy dry-run shows errors**
```bash
# Always test with dry-run first
./gforge deploy --dry-run

# Fix any validation errors before actual deployment
```

---

## 📖 Next Steps

- **Custom Domain**: Configure in Leapcell dashboard and Cloudflare
- **Monitor**: View logs at provider dashboards:
  - Leapcell: [leapcell.io/dashboard](https://leapcell.io/dashboard)
  - CockroachDB: [cockroachlabs.cloud](https://cockroachlabs.cloud)
  - Aiven: [console.aiven.io](https://console.aiven.io)
  - Cloudflare: [dash.cloudflare.com](https://dash.cloudflare.com)
- **Scale**: All providers offer generous free tiers with easy upgrades
- **Features**: Check `CONTRIBUTING.md` for adding features

---

## 🎓 Learning Resources

- **Full Docs**: `README.md`
- **Contributing**: `CONTRIBUTING.md`
- **Functions Guide**: `functions/README.md`
- **Provider Docs**:
  - [Leapcell Documentation](https://docs.leapcell.io)
  - [CockroachDB Serverless](https://www.cockroachlabs.com/docs/cockroachcloud/quickstart)
  - [Aiven Valkey](https://aiven.io/docs/products/valkey)
  - [Cloudflare Pages](https://developers.cloudflare.com/pages/)

---

## 📊 Comparison: Deployment Options

| Feature | Pages Only | Full Stack (Opinionated) |
|---------|-----------|--------------------------|
| **Setup Time** | 1 min | 5 min |
| **Deploy Command** | `deploy pages` | `deploy --with-valkey` |
| **Database** | ❌ | ✅ CockroachDB |
| **Backend** | ❌ | ✅ Leapcell |
| **Cache** | ❌ | ✅ Aiven Valkey |
| **Edge Functions** | ✅ | ✅ |
| **CDN** | ✅ | ✅ Cloudflare Proxy |
| **Cost (Free Tier)** | $0 | ~$0-10/month |
| **Best For** | Static sites | Production apps |

---

## 🏗️ Why This Stack?

Gothic Forge uses an **Opinionated Stack** to eliminate decision paralysis and provide a battle-tested production architecture:

### **Leapcell** (Compute)
- Docker-based deployments
- Automatic scaling
- Built-in health checks
- Simple pricing

### **CockroachDB Serverless** (Database)
- PostgreSQL-compatible
- Automatic scaling
- Built-in replication
- Generous free tier (5GB)

### **Aiven Valkey** (Cache)
- Redis-compatible
- Managed service
- High availability
- 30-day free trial

### **Cloudflare** (CDN & Proxy)
- Global edge network
- DDoS protection
- Automatic HTTPS
- Unlimited bandwidth (free tier)

This stack provides:
- ✅ **Global performance** - CDN + edge caching
- ✅ **Reliability** - Built-in redundancy and failover
- ✅ **Scalability** - Serverless architecture
- ✅ **Security** - DDoS protection, automatic HTTPS
- ✅ **Cost-effective** - Generous free tiers

---

## ⚡ Summary

Gothic Forge gets you from zero to production in **2 commands**:

```bash
# 1. Install
./gforge install

# 2. Deploy
./gforge deploy pages --run          # Static site (1 min)
# OR
./gforge deploy --with-valkey        # Full stack (5 min)
```

**That's it!** You're live with:
- ✅ Global CDN (Cloudflare)
- ✅ HTTPS by default
- ✅ Hot reload in dev
- ✅ Production-ready architecture (Opinionated Stack)
- ✅ Generous free tiers

**Now build something amazing!** 🚀
