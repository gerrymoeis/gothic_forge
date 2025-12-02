// Package jobs provides a background job processing system using Asynq.
//
// This package implements a Redis-backed job queue that supports:
//   - Asynchronous job processing
//   - Job scheduling with cron-like syntax
//   - Automatic retry with exponential backoff
//   - Job monitoring and inspection
//
// # Basic Usage
//
// Define a job by implementing the Job interface:
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
//
// Enqueue a job:
//
//	payload, _ := json.Marshal(EmailPayload{To: "user@example.com", Subject: "Hello"})
//	err := queue.Enqueue(ctx, "email:send", payload)
//
// # Requirements
//
// This package implements the following requirements:
//   - R5.2.1: Built-in job queue using Asynq (Redis-backed)
//   - R5.2.2: Job handler scaffolding support
//   - R5.2.3: Job scheduling with cron-like syntax
//   - R5.2.4: Job retry with exponential backoff
//   - R5.2.5: Web UI for job monitoring (via Asynq inspector)
package jobs
