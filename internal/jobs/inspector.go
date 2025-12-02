package jobs

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hibiken/asynq"
	"github.com/hibiken/asynqmon"
)

// InspectorConfig holds configuration for the job monitoring UI.
type InspectorConfig struct {
	// RedisAddr is the Redis server address (e.g., "localhost:6379")
	RedisAddr string

	// RedisPassword is the Redis password (empty string if no auth)
	RedisPassword string

	// RedisDB is the Redis database number (default: 0)
	RedisDB int

	// Port is the HTTP port for the monitoring UI (default: 8080)
	Port int
}

// StartInspector starts the Asynq monitoring web UI.
//
// The inspector provides a web interface to:
//   - View queues and their statistics
//   - List pending, active, scheduled, and failed jobs
//   - Inspect individual jobs
//   - Cancel or retry jobs
//   - View job history and metrics
//
// This method blocks until the server is stopped.
//
// Example:
//
//	config := InspectorConfig{
//	    RedisAddr: "localhost:6379",
//	    Port: 8080,
//	}
//	if err := StartInspector(config); err != nil {
//	    log.Fatal(err)
//	}
//
// Then visit http://localhost:8080 to view the monitoring UI.
func StartInspector(config InspectorConfig) error {
	if config.Port == 0 {
		config.Port = 8080
	}

	// Create Asynq inspector
	inspector := asynq.NewInspector(asynq.RedisClientOpt{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	// Create monitoring UI handler
	h := asynqmon.New(asynqmon.Options{
		RootPath:     "/",
		RedisConnOpt: asynq.RedisClientOpt{
			Addr:     config.RedisAddr,
			Password: config.RedisPassword,
			DB:       config.RedisDB,
		},
	})

	// Start HTTP server
	addr := fmt.Sprintf(":%d", config.Port)
	log.Printf("Starting job monitoring UI at http://localhost%s", addr)
	log.Printf("Redis: %s (DB: %d)", config.RedisAddr, config.RedisDB)

	// Store inspector for potential cleanup
	_ = inspector

	return http.ListenAndServe(addr, h)
}

// Inspector provides programmatic access to job queue inspection.
//
// Use this to build custom monitoring or management tools.
//
// Example:
//
//	inspector := NewInspector(config)
//	defer inspector.Close()
//
//	// List queues
//	queues, err := inspector.Queues()
//
//	// Get queue stats
//	stats, err := inspector.GetQueueInfo("default")
//
//	// List pending tasks
//	tasks, err := inspector.ListPendingTasks("default")
type Inspector struct {
	inspector *asynq.Inspector
}

// NewInspector creates a new job inspector.
func NewInspector(config InspectorConfig) *Inspector {
	inspector := asynq.NewInspector(asynq.RedisClientOpt{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	return &Inspector{
		inspector: inspector,
	}
}

// Queues returns a list of all queue names.
func (i *Inspector) Queues() ([]string, error) {
	return i.inspector.Queues()
}

// GetQueueInfo returns statistics for a specific queue.
func (i *Inspector) GetQueueInfo(queueName string) (*asynq.QueueInfo, error) {
	return i.inspector.GetQueueInfo(queueName)
}

// ListPendingTasks returns a list of pending tasks in a queue.
func (i *Inspector) ListPendingTasks(queueName string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error) {
	return i.inspector.ListPendingTasks(queueName, opts...)
}

// ListActiveTasks returns a list of active (currently processing) tasks in a queue.
func (i *Inspector) ListActiveTasks(queueName string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error) {
	return i.inspector.ListActiveTasks(queueName, opts...)
}

// ListScheduledTasks returns a list of scheduled tasks in a queue.
func (i *Inspector) ListScheduledTasks(queueName string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error) {
	return i.inspector.ListScheduledTasks(queueName, opts...)
}

// ListRetryTasks returns a list of tasks waiting to be retried.
func (i *Inspector) ListRetryTasks(queueName string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error) {
	return i.inspector.ListRetryTasks(queueName, opts...)
}

// ListArchivedTasks returns a list of archived (permanently failed) tasks.
func (i *Inspector) ListArchivedTasks(queueName string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error) {
	return i.inspector.ListArchivedTasks(queueName, opts...)
}

// CancelProcessing cancels a task that is currently being processed.
func (i *Inspector) CancelProcessing(taskID string) error {
	return i.inspector.CancelProcessing(taskID)
}

// DeleteTask deletes a task from the queue.
func (i *Inspector) DeleteTask(queueName, taskID string) error {
	return i.inspector.DeleteTask(queueName, taskID)
}

// RunTask runs a scheduled or retry task immediately.
func (i *Inspector) RunTask(queueName, taskID string) error {
	return i.inspector.RunTask(queueName, taskID)
}

// ArchiveTask archives a task (marks it as permanently failed).
func (i *Inspector) ArchiveTask(queueName, taskID string) error {
	return i.inspector.ArchiveTask(queueName, taskID)
}

// Close closes the inspector and releases resources.
func (i *Inspector) Close() error {
	return i.inspector.Close()
}
