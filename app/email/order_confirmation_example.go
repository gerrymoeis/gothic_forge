package email

import (
	"context"
	"log"
)

// ExampleOrderConfirmation demonstrates how to use the OrderConfirmation email template.
func ExampleOrderConfirmation() {
	ctx := context.Background()

	// Prepare email data
	data := OrderConfirmationData{
		// Fill in your data fields
		// Example: Name: "John Doe", VerificationURL: "https://example.com/verify/abc123"
	}

	// Send email
	err := SendOrderConfirmation(ctx, "user@example.com", data)
	if err != nil {
		log.Printf("Failed to send order-confirmation email: %v", err)
		return
	}

	log.Println("Order Confirmation email sent successfully")
}
