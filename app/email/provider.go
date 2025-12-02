package email

import (
	"fmt"
	"os"

	"gothicforge3/internal/email"
)

// initEmailProvider initializes the email provider based on environment variables.
// In development mode (APP_ENV=development), it automatically uses MailHog if available.
// Otherwise, it uses SMTP configuration.
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
			// If MailHog fails, fall through to SMTP
		}
	}

	// Use SMTP configuration
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
	return nil, fmt.Errorf("no email provider configured (set SMTP_HOST or start MailHog for development)")
}
