package email

import (
	"context"
	"fmt"
	"os"

	"gothicforge3/internal/email"
)

// initEmailProvider initializes the email provider based on environment variables.
// In development mode (APP_ENV=development), it automatically uses MailHog if available.
// Otherwise, it tries each provider in order: SendGrid, Mailgun, AWS SES, SMTP.
func initEmailProvider() (email.EmailProvider, error) {
	// In development mode, use MailHog if available
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" || appEnv == "development" {
		// Check if MailHog is explicitly enabled or if it's available
		useMailHog := os.Getenv("USE_MAILHOG")
		if useMailHog == "true" || useMailHog == "1" || (useMailHog == "" && email.IsMailHogAvailable()) {
			provider, err := email.NewMailHogProvider()
			if err == nil {
				return provider, nil
			}
			// If MailHog fails, fall through to other providers
		}
	}

	// Try to initialize based on available credentials
	if apiKey := os.Getenv("SENDGRID_API_KEY"); apiKey != "" {
		return email.NewSendGridProvider(apiKey)
	}
	if domain := os.Getenv("MAILGUN_DOMAIN"); domain != "" {
		apiKey := os.Getenv("MAILGUN_API_KEY")
		return email.NewMailgunProvider(domain, apiKey)
	}
	if region := os.Getenv("AWS_REGION"); region != "" {
		return email.NewSESProvider(context.Background(), region)
	}
	if host := os.Getenv("SMTP_HOST"); host != "" {
		port := 587
		if p := os.Getenv("SMTP_PORT"); p != "" {
			fmt.Sscanf(p, "%d", &port)
		}
		return email.NewSMTPProvider(email.SMTPConfig{
			Host:     host,
			Port:     port,
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
			UseTLS:   os.Getenv("SMTP_USE_TLS") != "false",
		})
	}
	return nil, fmt.Errorf("no email provider configured (set SENDGRID_API_KEY, MAILGUN_DOMAIN, AWS_REGION, SMTP_HOST, or start MailHog for development)")
}
