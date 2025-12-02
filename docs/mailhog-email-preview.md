# MailHog Email Preview in Development

Gothic Forge automatically integrates with MailHog for email preview in development mode. This allows you to test emails without sending them to real recipients.

## What is MailHog?

MailHog is an email testing tool that captures all outgoing emails and provides a web UI to view them. It's perfect for development and testing because:

- **No real emails sent**: All emails are captured locally
- **Web UI**: View emails in a browser at `http://localhost:8025`
- **Zero configuration**: Works automatically in development mode
- **Full email support**: HTML, text, attachments, multiple recipients

## Quick Start

### 1. Start MailHog

Using Docker (recommended):

```bash
docker run -d -p 1025:1025 -p 8025:8025 mailhog/mailhog
```

Or install MailHog directly:

```bash
# macOS
brew install mailhog
mailhog

# Linux
go install github.com/mailhog/MailHog@latest
MailHog

# Windows
# Download from https://github.com/mailhog/MailHog/releases
```

### 2. Set Development Mode

Make sure `APP_ENV` is set to `development` (or leave it unset):

```bash
export APP_ENV=development
```

### 3. Send Emails

Send emails normally in your application - they'll automatically be captured by MailHog:

```go
import "gothicforge3/app/email"

// Send a welcome email
data := email.WelcomeData{
    // Your data here
}
err := email.SendWelcome(ctx, "user@example.com", data)
```

### 4. View Emails

Open your browser to `http://localhost:8025` to view all captured emails.

## How It Works

In development mode (`APP_ENV=development` or unset), Gothic Forge automatically:

1. Checks if MailHog is running on `localhost:1025`
2. If available, routes all emails to MailHog instead of real email providers
3. Logs each email sent with a link to the MailHog UI

You'll see log messages like:

```
📧 MailHog email provider initialized (viewing at http://localhost:8025)
📧 Sending email to MailHog: noreply@example.com -> [user@example.com] (Subject: Welcome)
✓ Email sent to MailHog - view at http://localhost:8025
```

## Configuration

### Environment Variables

MailHog works with zero configuration, but you can customize it:

```bash
# Force MailHog usage (even in production - not recommended!)
USE_MAILHOG=true

# Custom MailHog host
MAILHOG_HOST=mailhog.local

# Custom MailHog port
MAILHOG_PORT=2025
```

### Disable MailHog in Development

To use real email providers in development:

```bash
# Set to production mode
export APP_ENV=production

# Or explicitly disable MailHog
export USE_MAILHOG=false
```

## Production Behavior

In production (`APP_ENV=production`), MailHog is automatically disabled and Gothic Forge uses real email providers (SendGrid, Mailgun, AWS SES, or SMTP).

## Docker Compose Example

Add MailHog to your `docker-compose.yml`:

```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=development
      - MAILHOG_HOST=mailhog
    depends_on:
      - mailhog

  mailhog:
    image: mailhog/mailhog
    ports:
      - "1025:1025"  # SMTP port
      - "8025:8025"  # Web UI port
```

## Testing with MailHog

MailHog is perfect for testing email functionality:

```go
func TestEmailSending(t *testing.T) {
    // Start MailHog before running tests
    // docker run -d -p 1025:1025 -p 8025:8025 mailhog/mailhog
    
    if !email.IsMailHogAvailable() {
        t.Skip("MailHog is not available")
    }
    
    // Send test email
    err := email.SendWelcome(context.Background(), "test@example.com", email.WelcomeData{})
    if err != nil {
        t.Fatalf("Failed to send email: %v", err)
    }
    
    // Check MailHog API to verify email was received
    // http://localhost:8025/api/v2/messages
}
```

## MailHog Web UI Features

The MailHog web UI at `http://localhost:8025` provides:

- **Email list**: View all captured emails
- **Email preview**: See HTML and text versions
- **Source view**: Inspect raw email source
- **Delete emails**: Clear the inbox
- **Search**: Find specific emails
- **Download**: Save emails as .eml files

## Troubleshooting

### Emails not appearing in MailHog

1. Check MailHog is running:
   ```bash
   curl http://localhost:8025
   ```

2. Check the logs for MailHog initialization:
   ```
   📧 MailHog email provider initialized (viewing at http://localhost:8025)
   ```

3. Verify APP_ENV is set to development:
   ```bash
   echo $APP_ENV
   ```

### MailHog not auto-detected

If MailHog is running but not being used:

```bash
# Force MailHog usage
export USE_MAILHOG=true
```

### Port conflicts

If port 1025 or 8025 is already in use:

```bash
# Use custom ports
docker run -d -p 2025:1025 -p 9025:8025 mailhog/mailhog

# Configure Gothic Forge
export MAILHOG_PORT=2025
```

## Alternative: Using SMTP Provider Directly

You can also configure the SMTP provider to use MailHog explicitly:

```bash
export SMTP_HOST=localhost
export SMTP_PORT=1025
export SMTP_USE_TLS=false
```

This works in any environment, not just development mode.

## Related Documentation

- [Email Package Documentation](../internal/email/README.md)
- [Email Scaffolding Guide](../app/email/SCAFFOLDING_GUIDE.md)
- [MailHog GitHub Repository](https://github.com/mailhog/MailHog)

## Summary

MailHog integration in Gothic Forge provides:

- ✅ **Zero configuration** - Works automatically in development
- ✅ **Safe testing** - No real emails sent
- ✅ **Visual preview** - See exactly what users will receive
- ✅ **Full features** - HTML, text, attachments, multiple recipients
- ✅ **Production-ready** - Automatically disabled in production

Start MailHog, set `APP_ENV=development`, and you're ready to test emails!
