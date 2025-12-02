// Package email provides email provider abstraction for Gothic Forge.
//
// The email package supports multiple email providers with a unified interface:
//   - MailHog (development/testing - automatic in dev mode)
//   - SendGrid (API-based)
//   - Mailgun (API-based)
//   - AWS SES (API-based)
//   - SMTP (protocol-based)
//
// # Usage
//
// Create a provider and send an email:
//
//	provider, err := email.NewSendGridProvider("your-api-key")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	email := &email.Email{
//	    From:    "sender@example.com",
//	    To:      []string{"recipient@example.com"},
//	    Subject: "Welcome to Gothic Forge",
//	    HTML:    "<h1>Welcome!</h1><p>Thanks for signing up.</p>",
//	    Text:    "Welcome! Thanks for signing up.",
//	}
//
//	err = provider.Send(context.Background(), email)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Provider Selection
//
// Choose a provider based on your needs:
//
//   - SendGrid: Easy to use, good deliverability, generous free tier
//   - Mailgun: Developer-friendly, powerful API, good for transactional emails
//   - AWS SES: Cost-effective at scale, integrates with AWS ecosystem
//   - SMTP: Universal protocol, works with any email server, good for self-hosted
//
// # Configuration
//
// Each provider requires different configuration:
//
// SendGrid:
//
//	provider, err := email.NewSendGridProvider(os.Getenv("SENDGRID_API_KEY"))
//
// Mailgun:
//
//	provider, err := email.NewMailgunProvider(
//	    os.Getenv("MAILGUN_DOMAIN"),
//	    os.Getenv("MAILGUN_API_KEY"),
//	)
//
// AWS SES:
//
//	provider, err := email.NewSESProvider(ctx, "us-east-1")
//
// SMTP:
//
//	provider, err := email.NewSMTPProvider(email.SMTPConfig{
//	    Host:     "smtp.example.com",
//	    Port:     587,
//	    Username: "user@example.com",
//	    Password: "password",
//	    UseTLS:   true,
//	})
//
// # Development & Testing
//
// Gothic Forge automatically uses MailHog for email preview in development mode.
// MailHog captures all outgoing emails and provides a web UI to view them.
//
// Quick Start:
//
//	# 1. Start MailHog with Docker
//	docker run -d -p 1025:1025 -p 8025:8025 mailhog/mailhog
//
//	# 2. Set APP_ENV to development (or leave unset)
//	export APP_ENV=development
//
//	# 3. Send emails normally - they'll be captured by MailHog
//	# View emails at http://localhost:8025
//
// Manual MailHog Provider:
//
//	provider, err := email.NewMailHogProvider()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	email := &email.Email{
//	    From:    "test@example.com",
//	    To:      []string{"user@example.com"},
//	    Subject: "Test Email",
//	    HTML:    "<h1>Hello from MailHog!</h1>",
//	    Text:    "Hello from MailHog!",
//	}
//
//	err = provider.Send(context.Background(), email)
//	// View at http://localhost:8025
//
// # Transactional Emails
//
// The package is designed for transactional emails (welcome, password reset, etc.).
// For marketing emails, consider using the provider's native features or a dedicated
// email marketing platform.
package email
