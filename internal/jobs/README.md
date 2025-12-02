# Jobs Package

The `jobs` package provides a background job processing system for Gothic Forge using [Asynq](https://github.com/hibiken/asynq).

## Features

- **Asynchronous Processing**: Jobs are processed in the background without blocking the main application
- **Redis-Backed**: Uses Redis for reliable job storage and distribution
- **Retry Logic**: Automatic retry with exponential backoff for failed jobs
- **Scheduling**: Support for delayed jobs and cron-like recurring jobs
- **Monitoring**: Built-in support for Asynq's web UI for job inspection
- **Priority Queues**: Process critical jobs before low-priority ones

## Requirements

This package implements the following requirements:

- **R5.2.1**: Built-in job queue using Asynq (Redis-backed)
- **R5.2.2**: Job handler scaffolding support
- **R5.2.3**: Job scheduling with cron-like syntax
- **R5.2.4**: Job retry with exponential backoff
- **R5.2.5**: Web UI for job monitoring

## Quick Start

### 1. Define a Job

Implement the `Job` interface:

```go
type SendEmailJob struct{}

func (j *SendEmailJob) Type() string {
    return "email:send"
}

func (j *SendEmailJob) Handle(ctx context.Context, payload []byte) error {
    var email EmailPayload
    if err := json.Unmarshal(payload, &email); err != nil {
        return err
    }
    return sendEmail(ctx, email)
}
```

### 2. Start a Worker

Create and start a worker to process jobs:

```go
config := jobs.WorkerConfig{
    RedisAddr: "localhost:6379",
    Concurrency: 10,
    Queues: map[string]int{
        "critical": 6,
        "default": 3,
        "low": 1,
    },
}

worker := jobs.NewWorker(config)
worker.Register(&SendEmailJob{})

go func() {
    if err := worker.Start(); err != nil {
        log.Fatal(err)
    }
}()
defer worker.Stop()
```

### 3. Enqueue Jobs

Create a queue client and enqueue jobs:

```go
queueConfig := jobs.QueueConfig{
    RedisAddr: "localhost:6379",
}
queue := jobs.NewQueue(queueConfig)
defer queue.Close()

// Enqueue a job
payload, _ := json.Marshal(EmailPayload{
    To: "user@example.com",
    Subject: "Welcome!",
    Body: "Thanks for signing up.",
})

err := queue.Enqueue(ctx, "email:send", payload)
```

## Advanced Usage

### Delayed Jobs

Schedule a job to run after a delay:

```go
// Send reminder in 24 hours
err := queue.EnqueueIn(ctx, "reminder:send", payload, 24*time.Hour)
```

### Scheduled Jobs

Schedule a job to run at a specific time:

```go
scheduledTime := time.Now().Add(24 * time.Hour)
err := queue.EnqueueAt(ctx, "reminder:send", payload, scheduledTime)
```

### Recurring Jobs

Use the scheduler for cron-like recurring jobs:

```go
schedulerConfig := jobs.SchedulerConfig{
    RedisAddr: "localhost:6379",
    Location: "UTC",
}
scheduler := jobs.NewScheduler(schedulerConfig)

// Run cleanup every day at 2 AM
entryID, err := scheduler.Register("0 2 * * *", "cleanup:old_data", nil)

go func() {
    if err := scheduler.Start(); err != nil {
        log.Fatal(err)
    }
}()
defer scheduler.Stop()
```

### Job Options

Customize job behavior with options:

```go
err := queue.Enqueue(ctx, "email:send", payload,
    asynq.MaxRetry(5),                    // Retry up to 5 times
    asynq.Timeout(30*time.Second),        // Timeout after 30 seconds
    asynq.Queue("critical"),              // Use critical queue
    asynq.ProcessIn(5*time.Minute),       // Delay by 5 minutes
)
```

### Priority Queues

Configure multiple queues with different priorities:

```go
config := jobs.WorkerConfig{
    RedisAddr: "localhost:6379",
    Concurrency: 10,
    Queues: map[string]int{
        "critical": 6,  // Highest priority
        "default": 3,   // Medium priority
        "low": 1,       // Lowest priority
    },
    StrictPriority: false, // Use weighted priority
}
```

### Custom Retry Logic

Provide custom retry delay logic:

```go
config := jobs.WorkerConfig{
    RedisAddr: "localhost:6379",
    RetryDelayFunc: func(n int, err error, task *asynq.Task) time.Duration {
        // Custom backoff: 5 * 2^n seconds
        return time.Duration(5*(1<<uint(n))) * time.Second
    },
}
```

## Monitoring

### Web UI

Gothic Forge includes a built-in web UI for monitoring jobs:

```bash
# Start the monitoring UI
gforge jobs

# Use a custom port
gforge jobs --port 3000
```

Then visit http://localhost:8080 to view:
- Queue statistics
- Pending, active, scheduled, and failed jobs
- Individual job details
- Job history and metrics

You can also cancel, retry, or archive jobs from the UI.

### Programmatic Access

For custom monitoring or management tools, use the Inspector:

```go
import "gothicforge3/internal/jobs"

config := jobs.InspectorConfig{
    RedisAddr: "localhost:6379",
}
inspector := jobs.NewInspector(config)
defer inspector.Close()

// List queues
queues, err := inspector.Queues()

// Get queue stats
stats, err := inspector.GetQueueInfo("default")

// List pending tasks
tasks, err := inspector.ListPendingTasks("default")

// Cancel a task
err = inspector.CancelProcessing(taskID)
```

### Asynq CLI

You can also use the official Asynq CLI:

```bash
# Install Asynq CLI
go install github.com/hibiken/asynq/tools/asynq@latest

# Monitor jobs
asynq stats

# View specific queue
asynq queue ls default

# Cancel a job
asynq task cancel <task-id>
```

## Example Jobs

The package includes several example jobs:

- `SendEmailJob`: Send an email notification
- `WebhookJob`: Send a webhook notification
- `CleanupJob`: Perform periodic cleanup
- `ReportJob`: Generate reports

See `examples.go` for implementation details.

## Scaffolding

Use the `gforge add job` command to scaffold new jobs:

```bash
# Create a new job
gforge add job SendWelcomeEmail

# This generates:
# - internal/jobs/send_welcome_email.go (job implementation)
# - internal/jobs/send_welcome_email_test.go (test file with basic tests)
```

The generated job includes:
- Job struct implementing the Job interface
- Payload struct for job data
- Type() method returning the job identifier
- Handle() method with TODO for implementation
- Complete test suite

Example generated code:

```go
type SendWelcomeEmail struct{}

type SendWelcomeEmailPayload struct {
    // Add your payload fields here
    UserID int64 `json:"user_id"`
    Email  string `json:"email"`
}

func (j *SendWelcomeEmail) Type() string {
    return "send:welcome:email"
}

func (j *SendWelcomeEmail) Handle(ctx context.Context, payload []byte) error {
    var data SendWelcomeEmailPayload
    if err := json.Unmarshal(payload, &data); err != nil {
        return fmt.Errorf("invalid payload: %w", err)
    }
    
    // TODO: Implement your job logic here
    return sendWelcomeEmail(ctx, data.UserID, data.Email)
}
```

## Testing

Test jobs by calling their `Handle()` method directly:

```go
func TestSendEmailJob(t *testing.T) {
    job := &SendEmailJob{}
    
    payload, _ := json.Marshal(EmailPayload{
        To: "test@example.com",
        Subject: "Test",
        Body: "Test body",
    })
    
    err := job.Handle(context.Background(), payload)
    assert.NoError(t, err)
}
```

For integration tests, use a test Redis instance:

```go
func TestJobQueue(t *testing.T) {
    // Use miniredis for testing
    mr, _ := miniredis.Run()
    defer mr.Close()
    
    config := jobs.QueueConfig{
        RedisAddr: mr.Addr(),
    }
    queue := jobs.NewQueue(config)
    defer queue.Close()
    
    // Test enqueueing
    err := queue.Enqueue(ctx, "test:job", []byte("test"))
    assert.NoError(t, err)
}
```

## Best Practices

1. **Keep Jobs Idempotent**: Jobs may be retried, so ensure they can be safely executed multiple times
2. **Use Timeouts**: Set reasonable timeouts to prevent jobs from running forever
3. **Handle Errors Gracefully**: Return errors for transient failures, panic for programming errors
4. **Log Important Events**: Use structured logging to track job execution
5. **Monitor Queue Depth**: Alert if queues grow too large
6. **Use Appropriate Queues**: Route jobs to the right priority queue
7. **Test Thoroughly**: Test both success and failure cases

## Architecture

```
┌─────────────┐
│   Client    │  Enqueues jobs
└──────┬──────┘
       │
       ▼
┌─────────────┐
│    Redis    │  Stores jobs
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Worker    │  Processes jobs
└─────────────┘
```

## Dependencies

- [Asynq](https://github.com/hibiken/asynq): Redis-backed job queue
- Redis: Job storage and distribution

## Related Packages

- `internal/email`: Email sending functionality
- `internal/storage`: File storage for job artifacts
- `internal/providers`: Provider integrations that may use jobs

## Complete Example

Here's a complete example showing how to set up and use the job system:

### 1. Create a Job

```bash
gforge add job SendWelcomeEmail
```

### 2. Implement the Job

Edit `internal/jobs/send_welcome_email.go`:

```go
type SendWelcomeEmailPayload struct {
    UserID int64  `json:"user_id"`
    Email  string `json:"email"`
    Name   string `json:"name"`
}

func (j *SendWelcomeEmail) Handle(ctx context.Context, payload []byte) error {
    var data SendWelcomeEmailPayload
    if err := json.Unmarshal(payload, &data); err != nil {
        return fmt.Errorf("invalid payload: %w", err)
    }
    
    // Send welcome email
    return emailService.Send(ctx, data.Email, "Welcome!", 
        fmt.Sprintf("Hello %s, welcome to our platform!", data.Name))
}
```

### 3. Start a Worker

Create `cmd/worker/main.go`:

```go
package main

import (
    "log"
    "gothicforge3/internal/jobs"
)

func main() {
    config := jobs.WorkerConfig{
        RedisAddr: "localhost:6379",
        Concurrency: 10,
        Queues: map[string]int{
            "critical": 6,
            "default": 3,
            "low": 1,
        },
    }
    
    worker := jobs.NewWorker(config)
    
    // Register jobs
    worker.Register(&jobs.SendWelcomeEmail{})
    worker.Register(&jobs.SendEmailJob{})
    worker.Register(&jobs.WebhookJob{})
    
    log.Println("Starting worker...")
    if err := worker.Start(); err != nil {
        log.Fatal(err)
    }
}
```

### 4. Enqueue Jobs

In your application:

```go
package main

import (
    "context"
    "encoding/json"
    "gothicforge3/internal/jobs"
)

func main() {
    config := jobs.QueueConfig{
        RedisAddr: "localhost:6379",
    }
    queue := jobs.NewQueue(config)
    defer queue.Close()
    
    // Enqueue a job
    payload, _ := json.Marshal(jobs.SendWelcomeEmailPayload{
        UserID: 123,
        Email: "user@example.com",
        Name: "John Doe",
    })
    
    err := queue.Enqueue(context.Background(), "send:welcome:email", payload)
    if err != nil {
        log.Fatal(err)
    }
}
```

### 5. Monitor Jobs

```bash
# Start the monitoring UI
gforge jobs

# Visit http://localhost:8080
```

## Future Enhancements

- Job result storage
- Job chaining (run job B after job A completes)
- Batch job processing
- Dead letter queue for permanently failed jobs
- Metrics and alerting integration
