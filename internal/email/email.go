// Package email provides email provider abstraction for Gothic Forge.
// It supports SMTP (universal standard) and MailHog (development testing)
// with a unified interface.
package email

import (
	"context"
	"fmt"
)

// EmailProvider defines the interface for sending emails.
// Implementations include SMTP (production) and MailHog (development).
type EmailProvider interface {
	// Send sends an email using the provider's API or protocol.
	Send(ctx context.Context, email *Email) error

	// Name returns the provider name (e.g., "smtp", "mailhog")
	Name() string
}

// Email represents an email message with all necessary fields.
type Email struct {
	// From is the sender's email address
	From string

	// To is a list of recipient email addresses
	To []string

	// Cc is a list of carbon copy recipients (optional)
	Cc []string

	// Bcc is a list of blind carbon copy recipients (optional)
	Bcc []string

	// Subject is the email subject line
	Subject string

	// HTML is the HTML version of the email body
	HTML string

	// Text is the plain text version of the email body
	Text string

	// ReplyTo is the reply-to email address (optional)
	ReplyTo string

	// Attachments is a list of file attachments (optional)
	Attachments []Attachment
}

// Attachment represents a file attachment for an email.
type Attachment struct {
	// Filename is the name of the file
	Filename string

	// Content is the file content as bytes
	Content []byte

	// ContentType is the MIME type (e.g., "application/pdf")
	ContentType string
}

// Validate checks if the email has all required fields.
func (e *Email) Validate() error {
	if e.From == "" {
		return fmt.Errorf("email from address is required")
	}
	if len(e.To) == 0 {
		return fmt.Errorf("email must have at least one recipient")
	}
	if e.Subject == "" {
		return fmt.Errorf("email subject is required")
	}
	if e.HTML == "" && e.Text == "" {
		return fmt.Errorf("email must have either HTML or text content")
	}
	return nil
}
