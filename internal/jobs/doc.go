// Package jobs provides minimal interfaces for background job processing.
//
// This package defines core interfaces without implementing a specific
// job queue system, allowing applications to choose their own implementation.
//
// # Interfaces
//
// Job - Represents a background task
// Queue - Enqueues jobs for processing
// Worker - Processes jobs from a queue
// Scheduler - Schedules jobs to run at specific times
//
// # Recommended Libraries
//
// For production use, consider:
//   - Asynq (https://github.com/hibiken/asynq) - Redis-based, recommended
//   - Machinery (https://github.com/RichardKnop/machinery) - Multiple backends
//   - River (https://github.com/riverqueue/river) - PostgreSQL-based
//
// For simple use cases:
//   - In-process channels
//   - PostgreSQL LISTEN/NOTIFY
//   - Cron jobs
//
// # Example with Asynq
//
//	import "github.com/hibiken/asynq"
//
//	// Define job
//	type EmailJob struct {
//	    To      string
//	    Subject string
//	    Body    string
//	}
//
//	func (j *EmailJob) Execute(ctx context.Context) error {
//	    // Send email
//	    return nil
//	}
//
//	// Enqueue
//	client := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})
//	task := asynq.NewTask("email:send", payload)
//	client.Enqueue(task)
//
// See README.md for complete examples and migration guide.
package jobs
