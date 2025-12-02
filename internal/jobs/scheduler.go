package jobs

import (
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// Scheduler manages periodic and scheduled jobs using cron-like syntax.
//
// The scheduler allows you to register jobs that run on a schedule,
// such as "every day at midnight" or "every 5 minutes".
//
// Example:
//
//	scheduler := NewScheduler(config)
//	defer scheduler.Stop()
//
//	// Run every day at midnight
//	scheduler.Register("0 0 * * *", "cleanup:old_data", nil)
//
//	// Run every 5 minutes
//	scheduler.Register("*/5 * * * *", "health:check", nil)
//
//	if err := scheduler.Start(); err != nil {
//	    log.Fatal(err)
//	}
type Scheduler struct {
	scheduler *asynq.Scheduler
	entries   map[string]string // entryID -> jobType
}

// SchedulerConfig holds configuration for the job scheduler.
type SchedulerConfig struct {
	// RedisAddr is the Redis server address (e.g., "localhost:6379")
	RedisAddr string

	// RedisPassword is the Redis password (empty string if no auth)
	RedisPassword string

	// RedisDB is the Redis database number (default: 0)
	RedisDB int

	// Location is the timezone for cron schedules (default: UTC)
	// Example: time.LoadLocation("America/New_York")
	Location string
}

// NewScheduler creates a new job scheduler.
//
// The scheduler uses cron syntax to schedule recurring jobs.
// Jobs are enqueued at the specified times and processed by workers.
//
// Example:
//
//	config := SchedulerConfig{
//	    RedisAddr: "localhost:6379",
//	    Location: "UTC",
//	}
//	scheduler := NewScheduler(config)
func NewScheduler(config SchedulerConfig) *Scheduler {
	opts := asynq.RedisClientOpt{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	}

	// Create scheduler with location
	var scheduler *asynq.Scheduler
	if config.Location != "" {
		loc, err := time.LoadLocation(config.Location)
		if err != nil {
			// Fall back to UTC if location is invalid
			loc = time.UTC
		}
		scheduler = asynq.NewScheduler(opts, &asynq.SchedulerOpts{
			Location: loc,
		})
	} else {
		scheduler = asynq.NewScheduler(opts, nil)
	}

	return &Scheduler{
		scheduler: scheduler,
		entries:   make(map[string]string),
	}
}

// Register registers a recurring job with a cron schedule.
//
// The cronspec uses standard cron syntax:
//   - "* * * * *" = every minute
//   - "0 * * * *" = every hour at minute 0
//   - "0 0 * * *" = every day at midnight
//   - "0 0 * * 0" = every Sunday at midnight
//   - "*/5 * * * *" = every 5 minutes
//
// The jobType should match a registered Job handler.
// The payload is optional data passed to the job.
//
// Returns an entry ID that can be used to unregister the job.
//
// Example:
//
//	// Run cleanup job every day at 2 AM
//	entryID, err := scheduler.Register("0 2 * * *", "cleanup:old_data", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (s *Scheduler) Register(cronspec, jobType string, payload []byte, opts ...asynq.Option) (string, error) {
	task := asynq.NewTask(jobType, payload, opts...)

	entryID, err := s.scheduler.Register(cronspec, task)
	if err != nil {
		return "", fmt.Errorf("failed to register scheduled job %s: %w", jobType, err)
	}

	s.entries[entryID] = jobType
	return entryID, nil
}

// Unregister removes a scheduled job.
//
// The entryID is returned by Register().
//
// Example:
//
//	err := scheduler.Unregister(entryID)
func (s *Scheduler) Unregister(entryID string) error {
	if err := s.scheduler.Unregister(entryID); err != nil {
		return fmt.Errorf("failed to unregister scheduled job: %w", err)
	}

	delete(s.entries, entryID)
	return nil
}

// Start starts the scheduler.
//
// This method blocks until Stop() is called or an error occurs.
// It should typically be run in a goroutine.
//
// Example:
//
//	go func() {
//	    if err := scheduler.Start(); err != nil {
//	        log.Fatal(err)
//	    }
//	}()
func (s *Scheduler) Start() error {
	if len(s.entries) == 0 {
		return fmt.Errorf("no scheduled jobs registered")
	}

	return s.scheduler.Run()
}

// Stop gracefully stops the scheduler.
func (s *Scheduler) Stop() {
	s.scheduler.Shutdown()
}
