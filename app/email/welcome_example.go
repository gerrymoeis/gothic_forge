package email

import (
	"context"
	"log"
)

// ExampleWelcome demonstrates how to use the Welcome email template.
func ExampleWelcome() {
	ctx := context.Background()

	// Prepare email data
	data := WelcomeData{
		// Fill in your data fields
		// Example: Name: "John Doe", VerificationURL: "https://example.com/verify/abc123"
	}

	// Send email
	err := SendWelcome(ctx, "user@example.com", data)
	if err != nil {
		log.Printf("Failed to send welcome email: %v", err)
		return
	}

	log.Println("Welcome email sent successfully")
}
