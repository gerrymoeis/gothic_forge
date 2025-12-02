package email

import (
	"context"
	"fmt"
	"os"

	"gothicforge3/internal/email"
)

// PasswordReset represents the data needed for the password-reset email.
type PasswordResetData struct {
	// Add your template data fields here
	// Example: Name string, VerificationURL string, etc.
}

// SendPasswordReset sends a password-reset email to the specified recipient.
//
// Example usage:
//
//	data := PasswordResetData{
//	    // Fill in your data
//	}
//	err := SendPasswordReset(ctx, "user@example.com", data)
func SendPasswordReset(ctx context.Context, to string, data PasswordResetData) error {
	// Initialize email provider
	provider, err := initEmailProvider()
	if err != nil {
		return fmt.Errorf("failed to initialize email provider: %w", err)
	}

	// Build email content
	subject := "Password Reset"
	htmlBody := buildPasswordResetHTML(data)
	textBody := buildPasswordResetText(data)

	// Create email
	msg := &email.Email{
		From:    os.Getenv("EMAIL_FROM"),
		To:      []string{to},
		Subject: subject,
		HTML:    htmlBody,
		Text:    textBody,
	}

	// Validate email
	if err := msg.Validate(); err != nil {
		return fmt.Errorf("invalid email: %w", err)
	}

	// Send email
	if err := provider.Send(ctx, msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// buildPasswordResetHTML generates the HTML version of the email.
func buildPasswordResetHTML(data PasswordResetData) string {
	// TODO: Implement your HTML email template
	// You can use html/template or a templating library
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Password Reset</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            text-align: center;
            border-radius: 8px 8px 0 0;
        }
        .content {
            background: #ffffff;
            padding: 30px;
            border: 1px solid #e0e0e0;
            border-top: none;
        }
        .footer {
            background: #f5f5f5;
            padding: 20px;
            text-align: center;
            font-size: 12px;
            color: #666;
            border-radius: 0 0 8px 8px;
        }
        .button {
            display: inline-block;
            padding: 12px 24px;
            background: #667eea;
            color: white;
            text-decoration: none;
            border-radius: 4px;
            margin: 20px 0;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Password Reset</h1>
    </div>
    <div class="content">
        <p>Hello,</p>
        <p>This is your password-reset email.</p>
        <!-- TODO: Add your email content here -->
        <p>Best regards,<br>Your Team</p>
    </div>
    <div class="footer">
        <p>&copy; 2024 Your Company. All rights reserved.</p>
    </div>
</body>
</html>
`
}

// buildPasswordResetText generates the plain text version of the email.
func buildPasswordResetText(data PasswordResetData) string {
	// TODO: Implement your plain text email template
	return `
Password Reset

Hello,

This is your password-reset email.

TODO: Add your email content here

Best regards,
Your Team

---
© 2024 Your Company. All rights reserved.
`
}
