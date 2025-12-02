package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// Example job implementations demonstrating the Job interface.
// These can be used as templates for creating new jobs.

// SendEmailJob is an example job that sends an email.
//
// This demonstrates a typical job pattern:
//   - Define a payload struct
//   - Unmarshal the payload in Handle()
//   - Perform the work
//   - Return an error if something goes wrong
//
// Example usage:
//
//	payload, _ := json.Marshal(EmailPayload{
//	    To: "user@example.com",
//	    Subject: "Welcome!",
//	    Body: "Thanks for signing up.",
//	})
//	queue.Enqueue(ctx, "email:send", payload)
type SendEmailJob struct{}

// EmailPayload contains the data needed to send an email.
type EmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// Type returns the job type identifier.
func (j *SendEmailJob) Type() string {
	return "email:send"
}

// Handle processes the email sending job.
func (j *SendEmailJob) Handle(ctx context.Context, payload []byte) error {
	var email EmailPayload
	if err := json.Unmarshal(payload, &email); err != nil {
		return fmt.Errorf("invalid email payload: %w", err)
	}

	// Validate required fields
	if email.To == "" {
		return fmt.Errorf("email recipient is required")
	}

	// Simulate sending email (replace with actual email service)
	log.Printf("Sending email to %s: %s", email.To, email.Subject)
	time.Sleep(100 * time.Millisecond) // Simulate network delay

	// In production, use an email service:
	// return emailService.Send(ctx, email)

	return nil
}

// WebhookJob is an example job that sends a webhook notification.
//
// Example usage:
//
//	payload, _ := json.Marshal(WebhookPayload{
//	    URL: "https://example.com/webhook",
//	    Event: "user.created",
//	    Data: map[string]interface{}{"user_id": 123},
//	})
//	queue.Enqueue(ctx, "webhook:notify", payload)
type WebhookJob struct{}

// WebhookPayload contains the data needed to send a webhook.
type WebhookPayload struct {
	URL   string                 `json:"url"`
	Event string                 `json:"event"`
	Data  map[string]interface{} `json:"data"`
}

// Type returns the job type identifier.
func (j *WebhookJob) Type() string {
	return "webhook:notify"
}

// Handle processes the webhook notification job.
func (j *WebhookJob) Handle(ctx context.Context, payload []byte) error {
	var webhook WebhookPayload
	if err := json.Unmarshal(payload, &webhook); err != nil {
		return fmt.Errorf("invalid webhook payload: %w", err)
	}

	// Validate required fields
	if webhook.URL == "" {
		return fmt.Errorf("webhook URL is required")
	}

	// Simulate sending webhook (replace with actual HTTP client)
	log.Printf("Sending webhook to %s: %s", webhook.URL, webhook.Event)
	time.Sleep(200 * time.Millisecond) // Simulate network delay

	// In production, use an HTTP client:
	// return httpClient.Post(ctx, webhook.URL, webhook.Data)

	return nil
}

// CleanupJob is an example job that performs periodic cleanup.
//
// This is typically scheduled to run periodically (e.g., daily).
//
// Example usage:
//
//	// Schedule to run every day at 2 AM
//	scheduler.Register("0 2 * * *", "cleanup:old_data", nil)
type CleanupJob struct{}

// Type returns the job type identifier.
func (j *CleanupJob) Type() string {
	return "cleanup:old_data"
}

// Handle processes the cleanup job.
func (j *CleanupJob) Handle(ctx context.Context, payload []byte) error {
	log.Println("Running cleanup job...")

	// Simulate cleanup work
	time.Sleep(500 * time.Millisecond)

	// In production, perform actual cleanup:
	// - Delete old records from database
	// - Remove expired cache entries
	// - Clean up temporary files
	// return database.DeleteOldRecords(ctx, 30*24*time.Hour)

	log.Println("Cleanup job completed")
	return nil
}

// ReportJob is an example job that generates a report.
//
// Example usage:
//
//	payload, _ := json.Marshal(ReportPayload{
//	    Type: "monthly_sales",
//	    Month: "2024-01",
//	})
//	queue.Enqueue(ctx, "report:generate", payload)
type ReportJob struct{}

// ReportPayload contains the data needed to generate a report.
type ReportPayload struct {
	Type  string `json:"type"`
	Month string `json:"month"`
}

// Type returns the job type identifier.
func (j *ReportJob) Type() string {
	return "report:generate"
}

// Handle processes the report generation job.
func (j *ReportJob) Handle(ctx context.Context, payload []byte) error {
	var report ReportPayload
	if err := json.Unmarshal(payload, &report); err != nil {
		return fmt.Errorf("invalid report payload: %w", err)
	}

	log.Printf("Generating %s report for %s", report.Type, report.Month)

	// Simulate report generation
	time.Sleep(1 * time.Second)

	// In production, generate actual report:
	// - Query database for data
	// - Generate PDF or CSV
	// - Upload to storage
	// - Send notification when complete

	log.Println("Report generation completed")
	return nil
}
