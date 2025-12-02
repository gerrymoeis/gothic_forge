package jobs

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
)

// Worker processes jobs from the queue.
//
// Worker runs in the background and processes jobs as they become available.
// Multiple workers can run concurrently to process jobs in parallel.
//
// Example:
//
//	worker := NewWorker(config)
//	worker.Register(&SendEmailJob{})
//	worker.Register(&WebhookJob{})
//
//	if err := worker.Start(); err != nil {
//	    log.Fatal(err)
//	}
//	defer worker.Stop()
type Worker struct {
	server   *asynq.Server
	mux      *asynq.ServeMux
	handlers map[string]Job
}

// WorkerConfig holds configuration for the job worker.
type WorkerConfig struct {
	// RedisAddr is the Redis server address (e.g., "localhost:6379")
	RedisAddr string

	// RedisPassword is the Redis password (empty string if no auth)
	RedisPassword string

	// RedisDB is the Redis database number (default: 0)
	RedisDB int

	// Concurrency is the number of concurrent workers (default: 10)
	Concurrency int

	// Queues maps queue names to priority levels.
	// Higher numbers = higher priority.
	// Example: map[string]int{"critical": 6, "default": 3, "low": 1}
	Queues map[string]int

	// StrictPriority ensures higher priority queues are processed first.
	// If false, queues are processed based on weighted priority.
	StrictPriority bool

	// RetryDelayFunc returns the delay duration before retrying a failed job.
	// If nil, uses exponential backoff: 2^retry seconds.
	RetryDelayFunc func(n int, err error, task *asynq.Task) time.Duration
}

// NewWorker creates a new job worker.
//
// The worker processes jobs from Redis using the provided configuration.
// Jobs must be registered using Register() before starting the worker.
//
// Example:
//
//	config := WorkerConfig{
//	    RedisAddr: "localhost:6379",
//	    Concurrency: 10,
//	    Queues: map[string]int{
//	        "critical": 6,
//	        "default": 3,
//	        "low": 1,
//	    },
//	}
//	worker := NewWorker(config)
func NewWorker(config WorkerConfig) *Worker {
	// Set defaults
	if config.Concurrency == 0 {
		config.Concurrency = 10
	}
	if config.Queues == nil {
		config.Queues = map[string]int{
			"default": 1,
		}
	}

	// Configure retry delay (exponential backoff)
	retryDelayFunc := config.RetryDelayFunc
	if retryDelayFunc == nil {
		retryDelayFunc = func(n int, err error, task *asynq.Task) time.Duration {
			// Exponential backoff: 2^n seconds
			// n=0: 1s, n=1: 2s, n=2: 4s, n=3: 8s, etc.
			delay := time.Duration(1<<uint(n)) * time.Second
			// Cap at 1 hour
			if delay > time.Hour {
				delay = time.Hour
			}
			return delay
		}
	}

	server := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     config.RedisAddr,
			Password: config.RedisPassword,
			DB:       config.RedisDB,
		},
		asynq.Config{
			Concurrency:    config.Concurrency,
			Queues:         config.Queues,
			StrictPriority: config.StrictPriority,
			RetryDelayFunc: retryDelayFunc,
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				// Log errors (in production, use structured logging)
				log.Printf("ERROR: job %s failed: %v", task.Type(), err)
			}),
		},
	)

	return &Worker{
		server:   server,
		mux:      asynq.NewServeMux(),
		handlers: make(map[string]Job),
	}
}

// Register registers a job handler with the worker.
//
// The job's Type() is used to route incoming tasks to the correct handler.
// Each job type can only be registered once.
//
// Example:
//
//	worker.Register(&SendEmailJob{})
//	worker.Register(&WebhookJob{})
func (w *Worker) Register(job Job) error {
	jobType := job.Type()

	if _, exists := w.handlers[jobType]; exists {
		return fmt.Errorf("job type %s is already registered", jobType)
	}

	w.handlers[jobType] = job

	// Create handler function that calls job.Handle()
	w.mux.HandleFunc(jobType, func(ctx context.Context, task *asynq.Task) error {
		return job.Handle(ctx, task.Payload())
	})

	return nil
}

// RegisterFunc registers a job handler function with the worker.
//
// This is a convenience method for simple handlers that don't need
// a full Job implementation.
//
// Example:
//
//	worker.RegisterFunc("simple:task", func(ctx context.Context, payload []byte) error {
//	    log.Printf("Processing: %s", payload)
//	    return nil
//	})
func (w *Worker) RegisterFunc(jobType string, handler func(context.Context, []byte) error) error {
	if _, exists := w.handlers[jobType]; exists {
		return fmt.Errorf("job type %s is already registered", jobType)
	}

	// Store a placeholder in handlers map
	w.handlers[jobType] = nil

	w.mux.HandleFunc(jobType, func(ctx context.Context, task *asynq.Task) error {
		return handler(ctx, task.Payload())
	})

	return nil
}

// Start starts the worker and begins processing jobs.
//
// This method blocks until Stop() is called or an error occurs.
// It should typically be run in a goroutine.
//
// Example:
//
//	go func() {
//	    if err := worker.Start(); err != nil {
//	        log.Fatal(err)
//	    }
//	}()
func (w *Worker) Start() error {
	if len(w.handlers) == 0 {
		return fmt.Errorf("no job handlers registered")
	}

	log.Printf("Starting job worker with %d registered handlers", len(w.handlers))
	return w.server.Run(w.mux)
}

// Stop gracefully stops the worker.
//
// This waits for currently processing jobs to complete before shutting down.
func (w *Worker) Stop() {
	w.server.Shutdown()
}
