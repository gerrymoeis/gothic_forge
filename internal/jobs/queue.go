package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// Queue manages job enqueueing and processing using Asynq.
//
// Queue provides methods to:
//   - Enqueue jobs for immediate processing
//   - Schedule jobs for future execution
//   - Schedule recurring jobs with cron syntax
//
// Example:
//
//	queue := NewQueue(redisAddr, redisPassword)
//	defer queue.Close()
//
//	// Enqueue a job
//	payload, _ := json.Marshal(EmailPayload{To: "user@example.com"})
//	err := queue.Enqueue(ctx, "email:send", payload)
type Queue struct {
	client *asynq.Client
}

// QueueConfig holds configuration for the job queue.
type QueueConfig struct {
	// RedisAddr is the Redis server address (e.g., "localhost:6379")
	RedisAddr string

	// RedisPassword is the Redis password (empty string if no auth)
	RedisPassword string

	// RedisDB is the Redis database number (default: 0)
	RedisDB int
}

// NewQueue creates a new job queue client.
//
// The queue client is used to enqueue jobs. It connects to Redis
// using the provided configuration.
//
// Example:
//
//	config := QueueConfig{
//	    RedisAddr: "localhost:6379",
//	    RedisPassword: "",
//	    RedisDB: 0,
//	}
//	queue := NewQueue(config)
//	defer queue.Close()
func NewQueue(config QueueConfig) *Queue {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	return &Queue{
		client: client,
	}
}

// Enqueue adds a job to the queue for immediate processing.
//
// The jobType should match the Type() of a registered Job handler.
// The payload is arbitrary data (typically JSON) that will be passed
// to the job's Handle() method.
//
// Options can be provided to customize job behavior:
//   - MaxRetry: maximum number of retry attempts
//   - Timeout: maximum time allowed for job execution
//   - Queue: name of the queue (for priority queues)
//
// Example:
//
//	payload, _ := json.Marshal(EmailPayload{To: "user@example.com"})
//	err := queue.Enqueue(ctx, "email:send", payload,
//	    asynq.MaxRetry(3),
//	    asynq.Timeout(30*time.Second),
//	)
func (q *Queue) Enqueue(ctx context.Context, jobType string, payload []byte, opts ...asynq.Option) error {
	task := asynq.NewTask(jobType, payload, opts...)
	info, err := q.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue job %s: %w", jobType, err)
	}

	// Log successful enqueue (in production, use structured logging)
	_ = info // info contains task ID, queue name, etc.

	return nil
}

// EnqueueIn schedules a job to be processed after the specified delay.
//
// This is useful for delayed jobs like "send reminder email in 24 hours".
//
// Example:
//
//	payload, _ := json.Marshal(ReminderPayload{UserID: 123})
//	err := queue.EnqueueIn(ctx, "reminder:send", payload, 24*time.Hour)
func (q *Queue) EnqueueIn(ctx context.Context, jobType string, payload []byte, delay time.Duration, opts ...asynq.Option) error {
	task := asynq.NewTask(jobType, payload, opts...)
	info, err := q.client.EnqueueContext(ctx, task, asynq.ProcessIn(delay))
	if err != nil {
		return fmt.Errorf("failed to schedule job %s: %w", jobType, err)
	}

	_ = info
	return nil
}

// EnqueueAt schedules a job to be processed at a specific time.
//
// Example:
//
//	scheduledTime := time.Now().Add(24 * time.Hour)
//	err := queue.EnqueueAt(ctx, "reminder:send", payload, scheduledTime)
func (q *Queue) EnqueueAt(ctx context.Context, jobType string, payload []byte, processAt time.Time, opts ...asynq.Option) error {
	task := asynq.NewTask(jobType, payload, opts...)
	info, err := q.client.EnqueueContext(ctx, task, asynq.ProcessAt(processAt))
	if err != nil {
		return fmt.Errorf("failed to schedule job %s: %w", jobType, err)
	}

	_ = info
	return nil
}

// Close closes the queue client and releases resources.
//
// This should be called when the application shuts down.
func (q *Queue) Close() error {
	return q.client.Close()
}
