package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// DemoMonitoringUI demonstrates the job monitoring UI functionality.
//
// This function:
// 1. Creates a queue and enqueues several jobs
// 2. Shows how to use the Inspector API to monitor jobs
// 3. Demonstrates the web UI capabilities
//
// To run this demo:
//
//	go run internal/jobs/monitoring_demo.go
//
// Then visit http://localhost:8080 to see the monitoring UI.
func DemoMonitoringUI() {
	fmt.Println("=== Job Monitoring UI Demo ===")
	fmt.Println()

	// Configuration
	config := QueueConfig{
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       0,
	}

	// Create queue
	queue := NewQueue(config)
	defer queue.Close()

	ctx := context.Background()

	// Enqueue some demo jobs
	fmt.Println("1. Enqueueing demo jobs...")

	// Immediate job
	payload1, _ := json.Marshal(EmailPayload{
		To:      "user1@example.com",
		Subject: "Welcome!",
		Body:    "Thanks for signing up.",
	})
	if err := queue.Enqueue(ctx, "email:send", payload1); err != nil {
		log.Printf("Error enqueueing job: %v", err)
	}
	fmt.Println("   ✓ Enqueued immediate email job")

	// Delayed job
	payload2, _ := json.Marshal(EmailPayload{
		To:      "user2@example.com",
		Subject: "Reminder",
		Body:    "Don't forget to complete your profile.",
	})
	if err := queue.EnqueueIn(ctx, "email:send", payload2, 5*time.Minute); err != nil {
		log.Printf("Error enqueueing delayed job: %v", err)
	}
	fmt.Println("   ✓ Enqueued delayed email job (5 minutes)")

	// Webhook job
	payload3, _ := json.Marshal(WebhookPayload{
		URL:   "https://example.com/webhook",
		Event: "user.created",
		Data:  map[string]interface{}{"user_id": 123},
	})
	if err := queue.Enqueue(ctx, "webhook:notify", payload3); err != nil {
		log.Printf("Error enqueueing webhook job: %v", err)
	}
	fmt.Println("   ✓ Enqueued webhook job")

	// Create inspector
	fmt.Println("\n2. Using Inspector API to monitor jobs...")
	inspectorConfig := InspectorConfig{
		RedisAddr:     config.RedisAddr,
		RedisPassword: config.RedisPassword,
		RedisDB:       config.RedisDB,
	}
	inspector := NewInspector(inspectorConfig)
	defer inspector.Close()

	// List queues
	queues, err := inspector.Queues()
	if err != nil {
		log.Printf("Error listing queues: %v", err)
	} else {
		fmt.Printf("   ✓ Found %d queue(s): %v\n", len(queues), queues)
	}

	// Get queue info
	if len(queues) > 0 {
		queueName := queues[0]
		info, err := inspector.GetQueueInfo(queueName)
		if err != nil {
			log.Printf("Error getting queue info: %v", err)
		} else {
			fmt.Printf("   ✓ Queue '%s' stats:\n", queueName)
			fmt.Printf("     - Pending: %d\n", info.Pending)
			fmt.Printf("     - Active: %d\n", info.Active)
			fmt.Printf("     - Scheduled: %d\n", info.Scheduled)
			fmt.Printf("     - Retry: %d\n", info.Retry)
			fmt.Printf("     - Archived: %d\n", info.Archived)
		}

		// List pending tasks
		tasks, err := inspector.ListPendingTasks(queueName)
		if err != nil {
			log.Printf("Error listing pending tasks: %v", err)
		} else {
			fmt.Printf("   ✓ Found %d pending task(s)\n", len(tasks))
			for i, task := range tasks {
				fmt.Printf("     %d. Type: %s, ID: %s\n", i+1, task.Type, task.ID)
			}
		}
	}

	// Start monitoring UI
	fmt.Println("\n3. Starting Web Monitoring UI...")
	fmt.Println("   → Visit http://localhost:8080 to view the monitoring UI")
	fmt.Println("   → Press Ctrl+C to stop")
	fmt.Println("\n   The UI provides:")
	fmt.Println("   - View queue statistics")
	fmt.Println("   - List pending, active, scheduled, and failed jobs")
	fmt.Println("   - Inspect individual jobs")
	fmt.Println("   - Cancel or retry jobs")
	fmt.Println("   - View job history and metrics")
	fmt.Println()

	// Start the inspector (this blocks)
	if err := StartInspector(inspectorConfig); err != nil {
		log.Fatalf("Error starting monitoring UI: %v", err)
	}
}

// DemoInspectorAPI demonstrates programmatic access to job monitoring.
//
// This shows how to build custom monitoring or management tools using
// the Inspector API without the web UI.
func DemoInspectorAPI() error {
	fmt.Println("=== Inspector API Demo ===")
	fmt.Println()

	config := InspectorConfig{
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       0,
	}

	inspector := NewInspector(config)
	defer inspector.Close()

	// List all queues
	fmt.Println("1. Listing all queues...")
	queues, err := inspector.Queues()
	if err != nil {
		return fmt.Errorf("failed to list queues: %w", err)
	}
	fmt.Printf("   Found %d queue(s): %v\n\n", len(queues), queues)

	// For each queue, show statistics
	for _, queueName := range queues {
		fmt.Printf("2. Queue: %s\n", queueName)

		// Get queue info
		info, err := inspector.GetQueueInfo(queueName)
		if err != nil {
			log.Printf("   Error getting queue info: %v", err)
			continue
		}

		fmt.Printf("   Statistics:\n")
		fmt.Printf("   - Pending: %d\n", info.Pending)
		fmt.Printf("   - Active: %d\n", info.Active)
		fmt.Printf("   - Scheduled: %d\n", info.Scheduled)
		fmt.Printf("   - Retry: %d\n", info.Retry)
		fmt.Printf("   - Archived: %d\n", info.Archived)
		fmt.Printf("   - Processed: %d\n", info.Processed)
		fmt.Printf("   - Failed: %d\n", info.Failed)

		// List pending tasks
		fmt.Println("\n   Pending tasks:")
		pendingTasks, err := inspector.ListPendingTasks(queueName)
		if err != nil {
			log.Printf("   Error listing pending tasks: %v", err)
		} else {
			if len(pendingTasks) == 0 {
				fmt.Println("   (none)")
			} else {
				for i, task := range pendingTasks {
					fmt.Printf("   %d. Type: %s, ID: %s, Retry: %d/%d\n",
						i+1, task.Type, task.ID, task.Retried, task.MaxRetry)
				}
			}
		}

		// List active tasks
		fmt.Println("\n   Active tasks:")
		activeTasks, err := inspector.ListActiveTasks(queueName)
		if err != nil {
			log.Printf("   Error listing active tasks: %v", err)
		} else {
			if len(activeTasks) == 0 {
				fmt.Println("   (none)")
			} else {
				for i, task := range activeTasks {
					fmt.Printf("   %d. Type: %s, ID: %s\n", i+1, task.Type, task.ID)
				}
			}
		}

		// List scheduled tasks
		fmt.Println("\n   Scheduled tasks:")
		scheduledTasks, err := inspector.ListScheduledTasks(queueName)
		if err != nil {
			log.Printf("   Error listing scheduled tasks: %v", err)
		} else {
			if len(scheduledTasks) == 0 {
				fmt.Println("   (none)")
			} else {
				for i, task := range scheduledTasks {
					fmt.Printf("   %d. Type: %s, ID: %s, Next run: %v\n",
						i+1, task.Type, task.ID, task.NextProcessAt)
				}
			}
		}

		fmt.Println()
	}

	return nil
}
