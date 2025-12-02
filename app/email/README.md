# Email Templates

This directory contains email templates for the application.

## Usage

Each email template is a Go package that provides:
- A data structure for template variables
- A Send function to send the email
- HTML and text versions of the email
- Tests and examples

## Sending Emails

```go
import "yourapp/app/email"

// Send an email
data := email.WelcomeData{
    Name: "John Doe",
}
err := email.SendWelcome(ctx, "user@example.com", data)
```

## Configuration

Set one of the following environment variables:

### SendGrid
```
SENDGRID_API_KEY=your-api-key
EMAIL_FROM=noreply@example.com
```

### Mailgun
```
MAILGUN_DOMAIN=mg.example.com
MAILGUN_API_KEY=your-api-key
EMAIL_FROM=noreply@example.com
```

### AWS SES
```
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
EMAIL_FROM=noreply@example.com
```

### SMTP
```
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_USE_TLS=true
EMAIL_FROM=noreply@example.com
```

## Testing

### Automatic MailHog Integration (Recommended)

Gothic Forge automatically uses MailHog for email preview in development mode:

```bash
# 1. Start MailHog
docker run -d -p 1025:1025 -p 8025:8025 mailhog/mailhog

# 2. Set development mode (or leave APP_ENV unset)
export APP_ENV=development
export EMAIL_FROM=test@example.com

# 3. Send emails - they'll automatically go to MailHog!
# View at http://localhost:8025
```

**No additional configuration needed!** In development mode, all emails are automatically captured by MailHog.

### Manual SMTP Configuration

Alternatively, configure SMTP explicitly:

```bash
export SMTP_HOST=localhost
export SMTP_PORT=1025
export EMAIL_FROM=test@example.com
```

For more details, see [MailHog Email Preview Documentation](../../docs/mailhog-email-preview.md)

## Adding New Templates

```bash
gforge add email <template-name> --provider=<provider>
```

Example:
```bash
gforge add email welcome --provider=sendgrid
gforge add email password-reset --provider=mailgun
```
