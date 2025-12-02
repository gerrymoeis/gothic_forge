# Email Template Scaffolding Guide

## Overview

The `gforge add email` command scaffolds complete email templates with provider-specific configuration, HTML/text versions, tests, and examples.

## Usage

```bash
gforge add email <template-name> [provider]
```

### Parameters

- `<template-name>`: Name of the email template (e.g., welcome, password-reset, verification)
- `[provider]`: Optional email provider (default: sendgrid)
  - Valid providers: `sendgrid`, `mailgun`, `ses`, `smtp`

### Examples

```bash
# Create a welcome email with SendGrid (default)
gforge add email welcome

# Create a password reset email with Mailgun
gforge add email password-reset mailgun

# Create a verification email with AWS SES
gforge add email verification ses

# Create a newsletter email with SMTP
gforge add email newsletter smtp
```

## Generated Files

For each email template, the following files are generated:

1. **`<template>.go`** - Main email template file with:
   - Data structure for template variables
   - `Send<Template>()` function to send the email
   - `build<Template>HTML()` function for HTML version
   - `build<Template>Text()` function for plain text version

2. **`<template>_test.go`** - Test file with:
   - Template generation tests
   - Email validation tests
   - Skips actual sending unless provider is configured

3. **`<template>_example.go`** - Example usage file demonstrating how to use the template

4. **`provider.go`** - Shared provider initialization (created once)

5. **`README.md`** - Documentation for the email package (created once)

## Configuration

Set the appropriate environment variables based on your provider:

### SendGrid
```env
SENDGRID_API_KEY=your-api-key
EMAIL_FROM=noreply@example.com
```

### Mailgun
```env
MAILGUN_DOMAIN=mg.example.com
MAILGUN_API_KEY=your-api-key
EMAIL_FROM=noreply@example.com
```

### AWS SES
```env
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
EMAIL_FROM=noreply@example.com
```

### SMTP
```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_USE_TLS=true
EMAIL_FROM=noreply@example.com
```

## Customization

After scaffolding, customize your email template:

1. **Add data fields** to the `<Template>Data` struct
2. **Customize HTML** in the `build<Template>HTML()` function
3. **Customize text** in the `build<Template>Text()` function
4. **Update subject** in the `Send<Template>()` function
5. **Add tests** in the `<template>_test.go` file

## Example: Welcome Email

```go
// 1. Define your data structure
type WelcomeData struct {
    Name           string
    VerificationURL string
}

// 2. Send the email
data := WelcomeData{
    Name:           "John Doe",
    VerificationURL: "https://example.com/verify/abc123",
}
err := SendWelcome(ctx, "user@example.com", data)
if err != nil {
    log.Printf("Failed to send welcome email: %v", err)
}
```

## Testing

Run tests for all email templates:

```bash
go test ./app/email/... -v
```

### Email Preview with MailHog (Automatic)

Gothic Forge automatically uses MailHog for email preview in development mode:

```bash
# 1. Start MailHog
docker run -d -p 1025:1025 -p 8025:8025 mailhog/mailhog

# 2. Set development mode (or leave APP_ENV unset)
export APP_ENV=development
export EMAIL_FROM=test@example.com

# 3. Run your application or tests
go test ./app/email/... -v

# 4. View emails at http://localhost:8025
```

**No additional configuration needed!** In development mode, the email system automatically detects MailHog and routes all emails to it for preview.

**Manual MailHog Configuration (Optional):**

```bash
# Force MailHog usage
export USE_MAILHOG=true

# Custom MailHog host/port
export MAILHOG_HOST=mailhog.local
export MAILHOG_PORT=2025
```

## Features

- ✅ Multiple provider support (SendGrid, Mailgun, AWS SES, SMTP)
- ✅ HTML and plain text versions
- ✅ Automatic provider detection from environment variables
- ✅ Email validation
- ✅ Context support for cancellation
- ✅ Comprehensive tests
- ✅ Example usage code
- ✅ Beautiful HTML templates with responsive design
- ✅ Error handling with descriptive messages

## Provider Auto-Detection

The `initEmailProvider()` function automatically detects which provider to use based on available environment variables, trying them in this order:

1. SendGrid (if `SENDGRID_API_KEY` is set)
2. Mailgun (if `MAILGUN_DOMAIN` is set)
3. AWS SES (if `AWS_REGION` is set)
4. SMTP (if `SMTP_HOST` is set)

This allows you to switch providers without changing code - just update your environment variables.

## Best Practices

1. **Always provide both HTML and text versions** - Some email clients only support plain text
2. **Test with MailHog locally** - Avoid sending real emails during development
3. **Use descriptive template names** - e.g., `welcome`, `password-reset`, `order-confirmation`
4. **Add all necessary data fields** - Include everything needed to render the email
5. **Keep templates simple** - Complex HTML may not render correctly in all email clients
6. **Test across email clients** - Gmail, Outlook, Apple Mail, etc.
7. **Monitor deliverability** - Check provider dashboards for bounce rates

## Troubleshooting

### "No email provider configured" error

Make sure you've set the appropriate environment variables for at least one provider.

### Emails not sending

1. Check your API keys are valid
2. Verify your sender email is authorized with the provider
3. Check provider dashboards for errors
4. Enable verbose logging to see detailed error messages

### HTML not rendering correctly

1. Use inline CSS (email clients don't support external stylesheets)
2. Test with email testing tools like Litmus or Email on Acid
3. Keep HTML simple and table-based for maximum compatibility

## Related Commands

- `gforge add job <name>` - Add background job (useful for async email sending)
- `gforge add auth` - Add authentication (often needs email verification)
- `gforge add api <name>` - Add API endpoint (for email webhooks)

## Support

For issues or questions:
- Check the main README.md
- Review internal/email/README.md for provider details
- See examples in app/email/*_example.go files
