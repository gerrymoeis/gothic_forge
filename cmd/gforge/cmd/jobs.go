package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gothicforge3/internal/env"
	"gothicforge3/internal/jobs"
)

var (
	jobsPort int
)

var jobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "Start the job monitoring UI",
	Long: `Start the Asynq job monitoring web interface.

The monitoring UI provides:
  - View queues and their statistics
  - List pending, active, scheduled, and failed jobs
  - Inspect individual jobs
  - Cancel or retry jobs
  - View job history and metrics

The UI will be available at http://localhost:8080 (or the port specified with --port).

Example:
  gforge jobs
  gforge jobs --port 3000`,
	RunE: func(cmd *cobra.Command, args []string) error {
		banner()

		// Load environment variables
		if err := env.Load(); err != nil {
			fmt.Printf("Warning: Could not load .env file: %v\n", err)
		}

		// Get Redis configuration from environment
		redisAddr := env.Get("REDIS_URL", "localhost:6379")
		redisPassword := env.Get("REDIS_PASSWORD", "")
		redisDB := 0 // Default to DB 0 for jobs

		// Parse Redis URL if it's a full URL
		if len(redisAddr) > 0 && redisAddr[:6] == "redis:" {
			// Simple parsing for redis://[:password@]host:port[/db]
			// For production, use a proper URL parser
			fmt.Println("Note: Using REDIS_URL. For custom configuration, set REDIS_ADDR, REDIS_PASSWORD, and REDIS_DB")
		}

		config := jobs.InspectorConfig{
			RedisAddr:     redisAddr,
			RedisPassword: redisPassword,
			RedisDB:       redisDB,
			Port:          jobsPort,
		}

		fmt.Println("Starting job monitoring UI...")
		fmt.Printf("Redis: %s (DB: %d)\n", redisAddr, redisDB)
		fmt.Printf("Web UI: http://localhost:%d\n", jobsPort)
		fmt.Println()
		fmt.Println("Press Ctrl+C to stop")

		if err := jobs.StartInspector(config); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting monitoring UI: %v\n", err)
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(jobsCmd)
	jobsCmd.Flags().IntVarP(&jobsPort, "port", "p", 8080, "Port for the monitoring UI")
}
