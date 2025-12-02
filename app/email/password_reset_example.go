package email

import (
	"context"
	"log"
)

// ExamplePasswordReset demonstrates how to use the PasswordReset email template.
func ExamplePasswordReset() {
	ctx := context.Background()

	// Prepare email data
	data := PasswordResetData{
		// Fill in your data fields
		// Example: Name: "John Doe", VerificationURL: "https://example.com/verify/abc123"
	}

	// Send email
	err := SendPasswordReset(ctx, "user@example.com", data)
	if err != nil {
		log.Printf("Failed to send password-reset email: %v", err)
		return
	}

	log.Println("Password Reset email sent successfully")
}
