# Job Monitoring UI Guide

## Overview

The Gothic Forge job monitoring UI provides a comprehensive web interface for monitoring and managing background jobs. Built on top of Asynq's monitoring capabilities, it offers real-time visibility into job queues, task execution, and system health.

## Features

### Web UI Features
- **Queue Statistics**: View real-time statistics for all job queues
- **Task Lists**: Browse pending, active, scheduled, retry, and archived tasks
- **Task Details**: Inspect individual tasks including payload, retry count, and execution history
- **Task Management**: Cancel, retry, or archive tasks directly from the UI
- **Job History**: View historical job execution data and metrics
- **Search & Filter**: Find specific jobs by type, status, or queue

### Programmatic Access
- **Inspector API**: Full programmatic access to job monitoring
- **Queue Management**: List queues and get detailed statistics
- **Task Operations**: Cancel, retry, delete, or archive tasks via code
- **Custom Monitoring**: Build custom dashboards and alerting systems

## Quick Start

### Starting the Monitoring UI

The simplest way to start the monitoring UI is with the CLI command:

```bash
# Start on default port (8080)
gforge jobs

# Start on custom port
gforge jobs --port 3000
```

Then visit `http://localhost:8080` (or your custom port) in your browser.

### Configuration

The monitoring UI reads Redis configuration from environment variables:

```bash
# .env file
REDIS_URL=localhost:6379
REDIS_PASSWORD=your-password-here
REDIS_DB=0
```

Or set them directly:

```bash
export REDIS_URL=localhost:6379
export REDIS_PASSWORD=secret
gforge jobs
```

## Using the Web UI

### Dashboard View

The main dashboard shows:
- **Queue Overview**: Number of tasks in each state (pending, active, scheduled, etc.)
- **Processing Rate**: Tasks processed per second/minute/hour
- **Error Rate**: Failed tasks and retry statistics
- **Queue Health**: Visual indicators for queue status

### Task Management

#### Viewing Tasks

1. Navigate to the queue you want to inspect
2. Select the task state (pending, active, scheduled, retry, archived)
3. Click on a task to view details

#### Task Details

Each task shows:
- **Task ID**: Unique identifier
- **Type**: Job type (e.g., `email:send`)
- **Payload**: JSON payload data
- **State**: Current state (pending, active, etc.)
- **Retry Count**: Number of retry attempts
- **Next Run**: Scheduled execution time (for scheduled tasks)
- **Error**: Error message (for failed tasks)

#### Task Actions

From the task details page, you can:
- **Cancel**: Stop a pending or scheduled task
- **Retry**: Immediately retry a failed task
- **Archive**: Move a task to the archive (permanent failure)
- **Delete**: Remove a task from the queue

### Queue Management

#### Queue Statistics

For each queue, view:
- **Size**: Number of tasks in the queue
- **Latency**: Average time tasks wait before processing
- **Processed**: Total tasks processed
- **Failed**: Total tasks that failed
- **Memory Usage**: Redis memory used by the queue

#### Queue Actions

- **Pause Queue**: Temporarily stop processing tasks
- **Resume Queue**: Resume processing paused tasks
- **Clear Queue**: Remove all tasks from a queue (use with caution!)

## Programmatic Access

### Using the Inspector API

The Inspector API provides programmatic access to all monitoring features:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "gothicforge3/internal/jobs"
)

func main() {
    // Create inspector
    config := jobs.InspectorConfig{
        RedisAddr:     "localhost:6379",
        RedisPassword: "",
        RedisDB:       0,
    }
    inspector := jobs.NewInspector(config)
    defer inspector.Close()

    // List all queues
    queues, err := inspector.Queues()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Found %d queues: %v\n", len(queues), queues)

    // Get queue statistics
    for _, queueName := range queues {
        info, err := inspector.GetQueueInfo(queueName)
        if err != nil {
            log.Printf("Error getting info for %s: %v", queueName, err)
            continue
        }

        fmt.Printf("\nQueue: %s\n", queueName)
        fmt.Printf("  Pending: %d\n", info.Pending)
        fmt.Printf("  Active: %d\n", info.Active)
        fmt.Printf("  Scheduled: %d\n", info.Scheduled)
        fmt.Printf("  Retry: %d\n", info.Retry)
        fmt.Printf("  Archived: %d\n", info.Archived)
        fmt.Printf("  Processed: %d\n", info.Processed)
        fmt.Printf("  Failed: %d\n", info.Failed)
    }
}
```

### Listing Tasks

```go
// List pending tasks
tasks, err := inspector.ListPendingTasks("default")
if err != nil {
    log.Fatal(err)
}

for _, task := range tasks {
    fmt.Printf("Task: %s (ID: %s)\n", task.Type, task.ID)
    fmt.Printf("  Retry: %d/%d\n", task.Retried, task.MaxRetry)
    fmt.Printf("  Payload: %s\n", string(task.Payload))
}
```

### Managing Tasks

```go
// Cancel a task
err := inspector.CancelProcessing(taskID)

// Delete a task
err := inspector.DeleteTask("default", taskID)

// Run a scheduled task immediately
err := inspector.RunTask("default", taskID)

// Archive a failed task
err := inspector.ArchiveTask("default", taskID)
```

### Building Custom Dashboards

```go
package main

import (
    "context"
    "time"
    "gothicforge3/internal/jobs"
)

func monitorQueues(ctx context.Context) {
    config := jobs.InspectorConfig{
        RedisAddr: "localhost:6379",
    }
    inspector := jobs.NewInspector(config)
    defer inspector.Close()

    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            queues, _ := inspector.Queues()
            for _, queueName := range queues {
                info, _ := inspector.GetQueueInfo(queueName)
                
                // Alert if queue is backing up
                if info.Pending > 1000 {
                    alertHighQueueDepth(queueName, info.Pending)
                }
                
                // Alert if error rate is high
                if info.Failed > info.Processed/10 {
                    alertHighErrorRate(queueName, info.Failed, info.Processed)
                }
            }
        }
    }
}
```

## Advanced Usage

### Custom Monitoring Server

You can embed the monitoring UI in your own HTTP server:

```go
package main

import (
    "net/http"
    "github.com/hibiken/asynq"
    "github.com/hibiken/asynqmon"
)

func main() {
    // Create monitoring handler
    h := asynqmon.New(asynqmon.Options{
        RootPath: "/monitoring",
        RedisConnOpt: asynq.RedisClientOpt{
            Addr: "localhost:6379",
        },
    })

    // Mount on custom path
    mux := http.NewServeMux()
    mux.Handle("/monitoring/", h)
    mux.HandleFunc("/", handleHome)

    http.ListenAndServe(":8080", mux)
}
```

### Authentication

Add authentication to the monitoring UI:

```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Check authentication
        username, password, ok := r.BasicAuth()
        if !ok || !checkCredentials(username, password) {
            w.Header().Set("WWW-Authenticate", `Basic realm="Job Monitoring"`)
            w.WriteHeader(401)
            w.Write([]byte("Unauthorized"))
            return
        }
        next.ServeHTTP(w, r)
    })
}

func main() {
    h := asynqmon.New(asynqmon.Options{
        RootPath: "/",
        RedisConnOpt: asynq.RedisClientOpt{
            Addr: "localhost:6379",
        },
    })

    http.ListenAndServe(":8080", authMiddleware(h))
}
```

### Metrics Export

Export metrics to Prometheus:

```go
package main

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "gothicforge3/internal/jobs"
)

var (
    queueSize = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "job_queue_size",
            Help: "Number of tasks in each queue",
        },
        []string{"queue", "state"},
    )
)

func init() {
    prometheus.MustRegister(queueSize)
}

func collectMetrics(inspector *jobs.Inspector) {
    queues, _ := inspector.Queues()
    for _, queueName := range queues {
        info, _ := inspector.GetQueueInfo(queueName)
        
        queueSize.WithLabelValues(queueName, "pending").Set(float64(info.Pending))
        queueSize.WithLabelValues(queueName, "active").Set(float64(info.Active))
        queueSize.WithLabelValues(queueName, "scheduled").Set(float64(info.Scheduled))
        queueSize.WithLabelValues(queueName, "retry").Set(float64(info.Retry))
        queueSize.WithLabelValues(queueName, "archived").Set(float64(info.Archived))
    }
}

func main() {
    // Start metrics collection
    go func() {
        config := jobs.InspectorConfig{RedisAddr: "localhost:6379"}
        inspector := jobs.NewInspector(config)
        defer inspector.Close()
        
        for {
            collectMetrics(inspector)
            time.Sleep(10 * time.Second)
        }
    }()

    // Expose metrics endpoint
    http.Handle("/metrics", promhttp.Handler())
    http.ListenAndServe(":9090", nil)
}
```

## Troubleshooting

### UI Not Loading

**Problem**: The monitoring UI doesn't load or shows connection errors.

**Solutions**:
1. Check Redis is running: `redis-cli ping`
2. Verify Redis connection: `redis-cli -h localhost -p 6379`
3. Check environment variables: `echo $REDIS_URL`
4. Verify port is not in use: `netstat -an | grep 8080`

### No Tasks Showing

**Problem**: The UI loads but shows no tasks.

**Solutions**:
1. Verify tasks are being enqueued: Check your application logs
2. Check Redis database: `redis-cli -n 0 KEYS "asynq:*"`
3. Verify queue names match: Check worker and queue configurations
4. Check Redis DB number: Ensure UI and workers use same DB

### Tasks Stuck in Pending

**Problem**: Tasks remain in pending state and never process.

**Solutions**:
1. Verify worker is running: Check worker process
2. Check worker logs: Look for errors or panics
3. Verify job handlers are registered: Check worker.Register() calls
4. Check queue priorities: Ensure worker is processing the right queues

### High Memory Usage

**Problem**: Redis memory usage is very high.

**Solutions**:
1. Check archived tasks: Archive old tasks to reduce memory
2. Set retention policy: Configure task retention in Asynq
3. Monitor queue depth: Alert on high queue sizes
4. Clean up old data: Periodically delete archived tasks

## Best Practices

### Security

1. **Restrict Access**: Don't expose the monitoring UI to the public internet
2. **Use Authentication**: Add authentication for production deployments
3. **Use HTTPS**: Serve the UI over HTTPS in production
4. **Limit Permissions**: Use read-only Redis credentials if possible

### Performance

1. **Use Pagination**: When listing tasks, use pagination for large queues
2. **Cache Statistics**: Cache queue statistics for dashboards
3. **Limit Polling**: Don't poll the API too frequently
4. **Use Indexes**: Ensure Redis has sufficient memory for indexes

### Monitoring

1. **Set Up Alerts**: Alert on high queue depth or error rates
2. **Track Metrics**: Export metrics to monitoring systems
3. **Log Important Events**: Log task failures and retries
4. **Review Regularly**: Regularly review archived tasks for patterns

### Maintenance

1. **Clean Up Archives**: Periodically delete old archived tasks
2. **Monitor Redis**: Keep an eye on Redis memory and performance
3. **Update Dependencies**: Keep Asynq and asynqmon up to date
4. **Backup Redis**: Regularly backup Redis data

## API Reference

### InspectorConfig

```go
type InspectorConfig struct {
    RedisAddr     string // Redis server address (e.g., "localhost:6379")
    RedisPassword string // Redis password (empty if no auth)
    RedisDB       int    // Redis database number (default: 0)
    Port          int    // HTTP port for monitoring UI (default: 8080)
}
```

### Inspector Methods

```go
// Queue operations
Queues() ([]string, error)
GetQueueInfo(queueName string) (*asynq.QueueInfo, error)

// Task listing
ListPendingTasks(queueName string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error)
ListActiveTasks(queueName string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error)
ListScheduledTasks(queueName string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error)
ListRetryTasks(queueName string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error)
ListArchivedTasks(queueName string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error)

// Task management
CancelProcessing(taskID string) error
DeleteTask(queueName, taskID string) error
RunTask(queueName, taskID string) error
ArchiveTask(queueName, taskID string) error

// Cleanup
Close() error
```

## Related Documentation

- [Job System README](README.md) - Complete job system documentation
- [Asynq Documentation](https://github.com/hibiken/asynq) - Asynq library docs
- [Asynqmon Documentation](https://github.com/hibiken/asynqmon) - Monitoring UI docs

## Support

For issues or questions:
1. Check the [troubleshooting section](#troubleshooting)
2. Review the [Asynq documentation](https://github.com/hibiken/asynq)
3. Open an issue on GitHub
4. Ask in the Gothic Forge community

## Conclusion

The job monitoring UI provides comprehensive visibility into your background job system. Use it to monitor job execution, debug issues, and ensure your job queues are healthy and performing well.
