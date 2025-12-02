# Gothic Forge - SaaS Template

A full-featured SaaS starter with authentication, user management, and subscription billing.

## What's Included

- ✅ User authentication (email/password + OAuth)
- ✅ User registration and login
- ✅ Email verification flow
- ✅ Password reset flow
- ✅ OAuth providers (GitHub, Google ready)
- ✅ User dashboard
- ✅ Subscription management
- ✅ Billing integration (Stripe ready)
- ✅ Role-based access control (RBAC)
- ✅ Session management with Redis
- ✅ Admin panel
- ✅ User profile management
- ✅ Team/organization support
- ✅ API key management

## Quick Start

```bash
# Install dependencies
gforge install

# Run migrations
gforge db --migrate

# Start development server
gforge dev

# Visit: http://localhost:8080
```

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go                  # Application entry point
├── internal/
│   ├── auth/
│   │   ├── handlers.go              # Auth HTTP handlers
│   │   ├── middleware.go            # Auth middleware
│   │   ├── jwt.go                   # JWT utilities
│   │   └── oauth.go                 # OAuth providers
│   ├── users/
│   │   ├── handlers.go              # User management
│   │   ├── models.go                # User models
│   │   └── repository.go            # Database operations
│   ├── billing/
│   │   ├── handlers.go              # Billing endpoints
│   │   ├── stripe.go                # Stripe integration
│   │   └── subscriptions.go         # Subscription logic
│   └── server/
│       └── server.go                # HTTP server setup
├── app/
│   ├── templates/
│   │   ├── layout.templ             # Base layout
│   │   ├── auth/
│   │   │   ├── login.templ          # Login page
│   │   │   ├── register.templ       # Registration page
│   │   │   └── reset.templ          # Password reset
│   │   ├── dashboard/
│   │   │   ├── home.templ           # User dashboard
│   │   │   ├── profile.templ        # User profile
│   │   │   └── settings.templ       # Account settings
│   │   └── admin/
│   │       ├── users.templ          # User management
│   │       └── analytics.templ      # Analytics dashboard
│   ├── static/
│   │   ├── styles/                  # CSS files
│   │   └── js/                      # JavaScript files
│   └── db/
│       └── migrations/
│           ├── 001_create_users.sql
│           ├── 002_create_sessions.sql
│           └── 003_create_subscriptions.sql
├── .env.example                     # Environment variables
├── providers.yaml                   # Provider configuration
└── go.mod                           # Go dependencies
```

## Database Schema

### Users Table

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    name VARCHAR(255),
    role VARCHAR(50) DEFAULT 'user',
    email_verified BOOLEAN DEFAULT FALSE,
    verification_token VARCHAR(255),
    reset_token VARCHAR(255),
    reset_token_expires TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Subscriptions Table

```sql
CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    stripe_customer_id VARCHAR(255),
    stripe_subscription_id VARCHAR(255),
    plan VARCHAR(50),
    status VARCHAR(50),
    current_period_start TIMESTAMP,
    current_period_end TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Available Endpoints

### Public Routes
- `GET /` - Homepage
- `GET /pricing` - Pricing page
- `GET /login` - Login page
- `POST /login` - Login handler
- `GET /register` - Registration page
- `POST /register` - Registration handler
- `GET /forgot-password` - Password reset request
- `POST /forgot-password` - Send reset email
- `GET /reset-password` - Password reset form
- `POST /reset-password` - Reset password handler

### OAuth Routes
- `GET /auth/github/login` - GitHub OAuth login
- `GET /auth/github/callback` - GitHub OAuth callback
- `GET /auth/google/login` - Google OAuth login
- `GET /auth/google/callback` - Google OAuth callback

### Protected Routes (Requires Authentication)
- `GET /dashboard` - User dashboard
- `GET /profile` - User profile
- `PUT /profile` - Update profile
- `GET /settings` - Account settings
- `POST /logout` - Logout handler

### Billing Routes (Requires Authentication)
- `GET /billing` - Billing dashboard
- `POST /billing/subscribe` - Create subscription
- `POST /billing/cancel` - Cancel subscription
- `POST /billing/update-payment` - Update payment method
- `GET /billing/invoices` - View invoices

### Admin Routes (Requires Admin Role)
- `GET /admin` - Admin dashboard
- `GET /admin/users` - User management
- `POST /admin/users/:id/role` - Update user role
- `DELETE /admin/users/:id` - Delete user

### API Routes
- `POST /api/v1/auth/token` - Get API token
- `GET /api/v1/me` - Get current user
- `GET /api/v1/users` - List users (admin only)

## Environment Variables

```bash
# Application
APP_ENV=development
SITE_BASE_URL=http://127.0.0.1:8080

# Server
HTTP_HOST=127.0.0.1
HTTP_PORT=8080

# Database (required)
DATABASE_URL=postgresql://user:pass@localhost:5432/saas

# Cache (required for sessions)
VALKEY_URL=redis://localhost:6379

# Security
JWT_SECRET=<auto-generated>

# OAuth (optional)
GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
OAUTH_BASE_URL=http://127.0.0.1:8080

# Email (required for verification/reset)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password
EMAIL_FROM=noreply@yourapp.com

# Stripe (required for billing)
STRIPE_SECRET_KEY=sk_test_...
STRIPE_PUBLISHABLE_KEY=pk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
```

## Features

### Authentication

- Email/password authentication with bcrypt
- OAuth integration (GitHub, Google)
- Email verification
- Password reset flow
- Session management with Redis
- JWT tokens for API access

### User Management

- User registration and login
- Profile management
- Role-based access control (user, admin)
- Account settings
- Email preferences

### Billing

- Stripe integration
- Subscription plans
- Payment method management
- Invoice history
- Webhook handling for subscription events

### Admin Panel

- User management
- Analytics dashboard
- Subscription overview
- System health monitoring

## Deployment

### Prerequisites

```bash
# Set up all required providers
gforge secrets --set COCKROACH_API_KEY=<your-key>
gforge secrets --set AIVEN_TOKEN=<your-token>
gforge secrets --set STRIPE_SECRET_KEY=<your-key>
gforge secrets --set GITHUB_CLIENT_ID=<your-id>
gforge secrets --set GITHUB_CLIENT_SECRET=<your-secret>
```

### Deploy

```bash
# Deploy with all services
gforge deploy --live --with-valkey
```

## Customization

### Add More OAuth Providers

Edit `internal/auth/oauth.go` to add providers like:
- Twitter/X
- LinkedIn
- Microsoft
- Apple

### Customize Email Templates

Edit templates in `internal/email/templates/`:
- Welcome email
- Verification email
- Password reset email
- Subscription notifications

### Add More Subscription Plans

Edit `internal/billing/plans.go` to define your pricing tiers.

## Next Steps

1. Configure OAuth providers in `.env`
2. Set up Stripe account and add keys
3. Configure email provider (SMTP or SendGrid)
4. Customize branding and colors
5. Deploy with `gforge deploy --live`

## Learn More

- [Gothic Forge Documentation](../../README.md)
- [Stripe Integration Guide](https://stripe.com/docs)
- [OAuth Setup Guide](https://oauth.net/2/)
- [Deployment Guide](../../QUICKSTART.md)
