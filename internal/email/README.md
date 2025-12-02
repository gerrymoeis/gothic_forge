# Email Package

The email package provides a simple, unified interface for sending emails in Gothic Forge.

## Philosophy

Gothic Forge takes a minimal approach to email:
- **SMTP** for production (universal standard, works with any provider)
- **MailHog** for development (email preview and testing)

This covers 99% of use cases while keeping dependencies minimal. If you need provider-specific features (SendGrid templates, Mailgun analytics, etc.), you can integrate them directly in your application.

## Supported Providers

- **SMTP** - Universal protocol, works with any email service (Gmail, Mailgun, SendGrid, AWS SES, etc.)
- **MailHog** - Development email testing with web UI preview

## Quick Start

### SMTP (Production)

SMTP works with any email service. Here are common configurations:

**Gmail:**
```go
provider, err := email.NewSMTPProvider(email.SMTPConfig{
    Host:     "smtp.gmail.com",
    Port:     587,
    Username: "your-email@gmail.com",
    Password: "your-app-password", // Use App Password, not regular password
    UseTLS:   true,
})
```

**Mailgun (via SMTP):**
```go
provider, err := email.NewSMTPProvider(email.SMTPConfig{
    Host:     "smtp.mailgun.org",
    Port:     587,
    Username: "postmaster@your-domain.com",
    Password: "your-mailgun-smtp-password",
    UseTLS:   true,
})
```

**SendGrid (via SMTP):**
```go
provider, err := email.NewSMTPProvider(email.SMTPConfig{
    Host:     "smtp.sendgrid.net",
    Port:     587,
    Username: "apikey", // Literally the string "apikey"
    Password: "your-sendgrid-api-key",
    UseTLS:   true,
})
```

**AWS SES (via SMTP):**
```go
provider, err := email.NewSMTPProvider(email.SMTPConfig{
    Host:     "email-smtp.us-east-1.amazonaws.com",
    Port:     587,
    Username: "your-ses-smtp-username",
    Password: "your-ses-smtp-password",
    UseTLS:   true,
})
```

**Sending an email:**
```go
email := &email.Email{
    From:    "noreply@example.com",
    To:      []string{"user@example.com"},
    Subject: "Welcome!",
    HTML:    "<h1>Welcome to Gothic Forge</h1>",
    Text:    "Welcome to Gothic Forge",
}

err = provider.Send(context.Background(), email)
if err != nil {
    log.Printf("Failed to send email: %v", err)
}
```

### MailHog (Development)

MailHog captures emails in development and provides a web UI to view them.

**Quick Start:**
```bash
# 1. Start MailHog with Docker
docker run -d -p 1025:1025 -p 8025:8025 mailhog/mailhog

# 2. Use MailHog provider in your code
```

**Code:**
```go
provider, err := email.NewMailHogProvider()
if err != nil {
    log.Fatal(err)
}

email := &email.Email{
    From:    "test@example.com",
    To:      []string{"user@example.com"},
    Subject: "Test Email",
    HTML:    "<h1>Hello from MailHog!</h1>",
    Text:    "Hello from MailHog!",
}

err = provider.Send(context.Background(), email)
// View at http://localhost:8025
```

**Custom Configuration:**
```go
provider, err := email.NewMailHogProvider(email.MailHogConfig{
    Host: "mailhog.local",
    Port: 2025,
})
```

## Email Structure

```go
type Email struct {
    From        string       // Sender email address
    To          []string     // Recipient email addresses
    Cc          []string     // Carbon copy recipients (optional)
    Bcc         []string     // Blind carbon copy recipients (optional)
    Subject     string       // Email subject
    HTML        string       // HTML version of email body
    Text        string       // Plain text version of email body
    ReplyTo     string       // Reply-to address (optional)
    Attachments []Attachment // File attachments (optional)
}
```

## Features

- **Universal SMTP**: Works with any email service
- **HTML & Text**: Support for both HTML and plain text emails
- **Attachments**: Send files with your emails
- **Multiple Recipients**: Send to multiple To, Cc, and Bcc addresses
- **Reply-To**: Set custom reply-to addresses
- **Validation**: Automatic validation of required fields
- **Context Support**: All operations support context for cancellation
- **Development Testing**: MailHog integration for email preview

## Environment Variables

```bash
# SMTP Configuration (Production)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_USE_TLS=true

# MailHog Configuration (Development)
MAILHOG_HOST=localhost  # Optional, defaults to localhost
MAILHOG_PORT=1025       # Optional, defaults to 1025
```

## Error Handling

All providers return descriptive errors:

```go
err := provider.Send(ctx, email)
if err != nil {
    // Handle specific errors
    switch {
    case strings.Contains(err.Error(), "authentication"):
        log.Println("Invalid credentials")
    case strings.Contains(err.Error(), "invalid email"):
        log.Println("Email validation failed")
    case strings.Contains(err.Error(), "connection"):
        log.Println("Cannot connect to SMTP server")
    default:
        log.Printf("Failed to send email: %v", err)
    }
}
```

## Best Practices

1. **Always provide both HTML and Text**: Some email clients only support plain text
2. **Validate emails before sending**: Use the `Validate()` method
3. **Use context with timeout**: Prevent hanging operations
4. **Handle errors gracefully**: Retry transient failures
5. **Test in development**: Use MailHog to preview emails
6. **Use App Passwords**: For Gmail, use App Passwords instead of your regular password
7. **Keep credentials secure**: Use environment variables, never hardcode

## Testing

```go
func TestEmailSending(t *testing.T) {
    // Use MailHog for testing
    provider, err := email.NewMailHogProvider()
    if err != nil {
        t.Skip("MailHog not available")
    }

    email := &email.Email{
        From:    "test@example.com",
        To:      []string{"user@example.com"},
        Subject: "Test",
        Text:    "Test email",
    }

    err = provider.Send(context.Background(), email)
    if err != nil {
        t.Fatalf("Failed to send email: %v", err)
    }
}
```

## Why SMTP Only?

**Advantages:**
- **Universal**: Works with any email service
- **No vendor lock-in**: Switch providers without code changes
- **Minimal dependencies**: No provider-specific SDKs
- **Simple**: One protocol to learn
- **Reliable**: Battle-tested standard

**When to use provider APIs directly:**
- You need provider-specific features (templates, analytics, webhooks)
- You're sending high volumes (>10k emails/day)
- You need advanced features (A/B testing, segmentation)

In these cases, integrate the provider SDK directly in your application rather than adding it to the framework core.

## Common SMTP Providers

| Provider | SMTP Host | Port | Free Tier |
|----------|-----------|------|-----------|
| Gmail | smtp.gmail.com | 587 | 500/day |
| Mailgun | smtp.mailgun.org | 587 | 5,000/month |
| SendGrid | smtp.sendgrid.net | 587 | 100/day |
| AWS SES | email-smtp.{region}.amazonaws.com | 587 | 62,000/month* |
| Postmark | smtp.postmarkapp.com | 587 | 100/month |
| Brevo (Sendinblue) | smtp-relay.brevo.com | 587 | 300/day |

*AWS SES free tier requires EC2 instance

## Migration from Provider-Specific Code

If you were using SendGrid, Mailgun, or AWS SES APIs directly, migration is simple:

**Before (SendGrid API):**
```go
provider, err := email.NewSendGridProvider(apiKey)
```

**After (SendGrid via SMTP):**
```go
provider, err := email.NewSMTPProvider(email.SMTPConfig{
    Host:     "smtp.sendgrid.net",
    Port:     587,
    Username: "apikey",
    Password: apiKey,
    UseTLS:   true,
})
```

The rest of your code remains unchanged!

## Future Enhancements

- Email template system
- Batch sending
- Retry logic with exponential backoff
- Email queue integration

## Need More Features?

If you need provider-specific features (templates, analytics, webhooks), consider:

1. **Use provider's SMTP** for basic sending (covered here)
2. **Integrate provider SDK directly** in your application for advanced features
3. **Create a plugin/extension** if you want to share with the community

This keeps the framework minimal while allowing flexibility for advanced use cases.
