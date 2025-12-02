package email

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

// MailHogProvider is a development-only email provider that sends emails to MailHog.
// MailHog is an email testing tool that captures emails and provides a web UI to view them.
// This provider uses SMTP to send emails to MailHog's SMTP server.
//
// Default configuration:
//   - Host: localhost
//   - Port: 1025
//   - No authentication required
//
// To start MailHog:
//   docker run -d -p 1025:1025 -p 8025:8025 mailhog/mailhog
//
// View emails at: http://localhost:8025
type MailHogProvider struct {
	smtp *SMTPProvider
	host string
	port int
}

// NewMailHogProvider creates a new MailHog email provider for development.
// It uses the SMTP provider internally to send emails to MailHog.
//
// Configuration via environment variables:
//   - MAILHOG_HOST: MailHog SMTP host (default: localhost)
//   - MAILHOG_PORT: MailHog SMTP port (default: 1025)
//
// Example:
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
func NewMailHogProvider() (*MailHogProvider, error) {
	host := os.Getenv("MAILHOG_HOST")
	if host == "" {
		host = "localhost"
	}

	port := 1025
	if portStr := os.Getenv("MAILHOG_PORT"); portStr != "" {
		fmt.Sscanf(portStr, "%d", &port)
	}

	// Create SMTP provider for MailHog
	smtpConfig := SMTPConfig{
		Host:     host,
		Port:     port,
		Username: "", // MailHog doesn't require authentication
		Password: "",
		UseTLS:   false, // MailHog doesn't use TLS by default
	}

	smtp, err := NewSMTPProvider(smtpConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create MailHog SMTP provider: %w", err)
	}

	log.Printf("📧 MailHog email provider initialized (viewing at http://%s:8025)", host)

	return &MailHogProvider{
		smtp: smtp,
		host: host,
		port: port,
	}, nil
}

// Send sends an email to MailHog for preview.
// The email will be captured by MailHog and can be viewed in the web UI.
func (p *MailHogProvider) Send(ctx context.Context, email *Email) error {
	if err := email.Validate(); err != nil {
		return fmt.Errorf("mailhog: email validation failed: %w", err)
	}

	// Log that we're sending to MailHog for visibility
	log.Printf("📧 Sending email to MailHog: %s -> %v (Subject: %s)", email.From, email.To, email.Subject)

	// Use the SMTP provider to send to MailHog
	if err := p.smtp.Send(ctx, email); err != nil {
		return fmt.Errorf("mailhog: failed to send email: %w", err)
	}

	log.Printf("✓ Email sent to MailHog - view at http://%s:8025", p.host)
	return nil
}

// Name returns the provider name.
func (p *MailHogProvider) Name() string {
	return "mailhog"
}

// IsMailHogAvailable checks if MailHog is running and available.
// It attempts to connect to the MailHog SMTP port.
func IsMailHogAvailable() bool {
	host := os.Getenv("MAILHOG_HOST")
	if host == "" {
		host = "localhost"
	}

	port := 1025
	if portStr := os.Getenv("MAILHOG_PORT"); portStr != "" {
		fmt.Sscanf(portStr, "%d", &port)
	}

	// Try to actually connect to MailHog
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
