# Jobs Package

The jobs package provides minimal interfaces for background job processing in Gothic Forge.

## Philosophy

Gothic Forge takes a minimal approach to background jobs:
- **Interfaces only** - No specific implementation
- **Flexibility** - Choose your own job queue system
- **Simplicity** - Keep the core framework lightweight

This allows you to use any job queue library (asynq, machinery, etc.) or implement your own simple queue without framework lock-in.

## Interfaces

### Job
```go
type Job interface {
    Execute(ctx context.Context) error
}
```

Represents a background task that can be executed.

### Queue
```go
type Queue interface {
    Enqueue(ctx context.Context, job Job) error
}
```

Represents a job queue that can enqueue jobs for processing.

### Worker
```go
type Worker interface {
    Start(ctx context.Context) error
    Stop() error
}
```

Represents a worker that processes jobs from a queue.

### Scheduler
```go
type Scheduler interface {
    Schedule(ctx context.Context, job Job, schedule string) error
    Start(ctx context.Context) error
    Stop() error
}
```

Represents a scheduler for running jobs at specific times or intervals.

## Usage

### Option 1: Use a Job Queue Library

**Asynq (Redis-based, recommended):**
```go
import "github.com/hibiken/asynq"

// Define your job
type EmailJob struct {
    To      string
    Subject string
    Body    string
}

func (j *EmailJob) Execute(ctx context.Context) error {
    // Send email logic
    return nil
}

// Create client
client := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})
defer client.Close()

// Enqueue job
task := asynq.NewTask("email:send", payload)
client.Enqueue(task)
```

**Machinery (Multiple backends):**
```go
import "github.com/RichardKnop/machinery/v1"

server, err := machinery.NewServer(config)
server.RegisterTask("send_email", SendEmail)
```

### Option 2: Simple In-Process Queue

For simple use cases, implement a basic in-process queue:

```go
type SimpleQueue struct {
    jobs chan jobs.Job
}

func NewSimpleQueue(bufferSize int) *SimpleQueue {
    return &SimpleQueue{
        jobs: make(chan jobs.Job, bufferSize),
    }
}

func (q *SimpleQueue) Enqueue(ctx context.Context, job jobs.Job) error {
    select {
    case q.jobs <- job:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

func (q *SimpleQueue) Start(ctx context.Context) error {
    for {
        select {
        case job := <-q.jobs:
            if err := job.Execute(ctx); err != nil {
                log.Printf("Job failed: %v", err)
            }
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}
```

### Option 3: Database-Backed Queue

For persistence without Redis:

```go
// Use PostgreSQL LISTEN/NOTIFY
// Or use a jobs table with polling
type DBQueue struct {
    db *sql.DB
}

func (q *DBQueue) Enqueue(ctx context.Context, job jobs.Job) error {
    // Insert job into database
    _, err := q.db.ExecContext(ctx, 
        "INSERT INTO jobs (type, payload, status) VALUES ($1, $2, 'pending')",
        job.Type(), job.Payload())
    return err
}
```

## Recommended Libraries

### Production Use:

1. **[Asynq](https://github.com/hibiken/asynq)** (Recommended)
   - Redis-based
   - Reliable, battle-tested
   - Built-in monitoring UI
   - Retry logic, scheduling
   - ~10k stars

2. **[Machinery](https://github.com/RichardKnop/machinery)**
   - Multiple backends (Redis, AMQP, etc.)
   - Flexible
   - ~7k stars

3. **[River](https://github.com/riverqueue/river)**
   - PostgreSQL-based
   - No Redis required
   - Modern, type-safe
   - ~3k stars

### Simple Use Cases:

1. **In-process channels** - For simple async tasks
2. **PostgreSQL LISTEN/NOTIFY** - For database-backed queues
3. **Cron jobs** - For scheduled tasks only

## Example: Integrating Asynq

### 1. Install Asynq
```bash
go get github.com/hibiken/asynq
```

### 2. Define Jobs
```go
// app/jobs/email.go
package jobs

import (
    "context"
    "encoding/json"
    "github.com/hibiken/asynq"
)

type EmailPayload struct {
    To      string
    Subject string
    Body    string
}

func NewEmailTask(payload EmailPayload) (*asynq.Task, error) {
    data, err := json.Marshal(payload)
    if err != nil {
        return nil, err
    }
    return asynq.NewTask("email:send", data), nil
}

func HandleEmailTask(ctx context.Context, t *asynq.Task) error {
    var p EmailPayload
    if err := json.Unmarshal(t.Payload(), &p); err != nil {
        return err
    }
    // Send email
    return nil
}
```

### 3. Enqueue Jobs
```go
// In your handlers
client := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})
defer client.Close()

task, err := jobs.NewEmailTask(jobs.EmailPayload{
    To:      "user@example.com",
    Subject: "Welcome",
    Body:    "Welcome to our app!",
})
if err != nil {
    return err
}

info, err := client.Enqueue(task)
```

### 4. Process Jobs
```go
// cmd/worker/main.go
package main

import (
    "github.com/hibiken/asynq"
    "yourapp/app/jobs"
)

func main() {
    srv := asynq.NewServer(
        asynq.RedisClientOpt{Addr: "localhost:6379"},
        asynq.Config{Concurrency: 10},
    )

    mux := asynq.NewServeMux()
    mux.HandleFunc("email:send", jobs.HandleEmailTask)

    if err := srv.Run(mux); err != nil {
        log.Fatal(err)
    }
}
```

## Why No Built-in Implementation?

**Reasons:**
1. **Flexibility** - Different apps need different solutions
2. **Minimal dependencies** - Keep the framework lightweight
3. **No lock-in** - Choose the best tool for your needs
4. **Simplicity** - Interfaces are easier to understand than implementations

**When you need jobs:**
- Add the library that fits your needs
- Implement the interfaces
- Keep it in your application code, not the framework

## Migration from v9.2

If you were using the built-in jobs system in v9.2:

**Before:**
```go
import "gothicforge3/internal/jobs"

queue := jobs.NewQueue(redisOpt)
queue.Enqueue(ctx, &jobs.EmailJob{...})
```

**After (with Asynq):**
```go
import "github.com/hibiken/asynq"

client := asynq.NewClient(redisOpt)
task := asynq.NewTask("email:send", payload)
client.Enqueue(task)
```

The migration is straightforward - just replace the framework's job queue with Asynq directly.

## Best Practices

1. **Use a proven library** - Don't reinvent the wheel for production
2. **Handle failures gracefully** - Implement retry logic
3. **Monitor your queues** - Use monitoring tools (Asynq UI, etc.)
4. **Keep jobs idempotent** - Jobs should be safe to retry
5. **Use timeouts** - Always use context with timeouts
6. **Log everything** - Jobs run in background, logging is crucial

## Resources

- [Asynq Documentation](https://github.com/hibiken/asynq/wiki)
- [Machinery Documentation](https://github.com/RichardKnop/machinery)
- [River Documentation](https://riverqueue.com/docs)
- [Background Jobs Best Practices](https://github.com/bensheldon/good_job#readme)

## Need Help?

If you need help choosing or implementing a job queue:
1. Check the recommended libraries above
2. Review their documentation
3. Start with Asynq if unsure (most popular, well-documented)
4. For simple cases, use in-process channels

This keeps Gothic Forge minimal while giving you full flexibility for background jobs!
