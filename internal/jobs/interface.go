// Package jobs provides a minimal interface for background job processing.
//
// This package defines the core interfaces for job processing without
// implementing a specific job queue system. This allows applications to
// choose their own job queue implementation (asynq, machinery, etc.) or
// implement a simple in-process queue.
//
// For a full-featured job queue implementation, see:
// https://github.com/hibiken/asynq (Redis-based)
// https://github.com/RichardKnop/machinery (Multiple backends)
package jobs

import "context"

// Job represents a background task that can be executed.
type Job interface {
	// Execute runs the job with the given context.
	// It should return an error if the job fails.
	Execute(ctx context.Context) error
}

// Queue represents a job queue that can enqueue jobs for processing.
type Queue interface {
	// Enqueue adds a job to the queue for processing.
	Enqueue(ctx context.Context, job Job) error
}

// Worker represents a worker that processes jobs from a queue.
type Worker interface {
	// Start begins processing jobs from the queue.
	Start(ctx context.Context) error

	// Stop gracefully stops the worker.
	Stop() error
}

// Scheduler represents a scheduler that can schedule jobs to run at specific times.
type Scheduler interface {
	// Schedule schedules a job to run at a specific time or interval.
	Schedule(ctx context.Context, job Job, schedule string) error

	// Start begins the scheduler.
	Start(ctx context.Context) error

	// Stop gracefully stops the scheduler.
	Stop() error
}
