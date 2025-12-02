package jobs

import (
	"context"
)

// Job represents a background job that can be processed asynchronously.
//
// Jobs are identified by their Type() and process arbitrary payloads
// through their Handle() method. The payload is typically JSON-encoded
// data that contains all information needed to execute the job.
//
// Example implementation:
//
//	type SendEmailJob struct{}
//
//	func (j *SendEmailJob) Type() string {
//	    return "email:send"
//	}
//
//	func (j *SendEmailJob) Handle(ctx context.Context, payload []byte) error {
//	    var email EmailPayload
//	    if err := json.Unmarshal(payload, &email); err != nil {
//	        return err
//	    }
//	    return sendEmail(ctx, email)
//	}
type Job interface {
	// Type returns the unique identifier for this job type.
	// This is used to route jobs to the correct handler.
	// Convention: use colon-separated namespaces (e.g., "email:send", "webhook:notify")
	Type() string

	// Handle processes the job with the given payload.
	// The payload is typically JSON-encoded data.
	//
	// If Handle returns an error, the job will be retried according to
	// the configured retry policy. Return nil to mark the job as successful.
	//
	// The context may be cancelled if the job exceeds its timeout or
	// if the worker is shutting down.
	Handle(ctx context.Context, payload []byte) error
}
