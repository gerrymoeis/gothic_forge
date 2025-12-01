package cmd

import (
  "fmt"
  "github.com/spf13/cobra"
)

var (
  logsFollow bool
  logsSince  string
)

var logsCmd = &cobra.Command{
  Use:   "logs",
  Short: "View application logs",
  Long: `View application logs for your deployed Gothic Forge app.

Gothic Forge uses Leapcell for backend compute.
View your application logs in the Leapcell dashboard:

  1. Log in to https://leapcell.io
  2. Select your app
  3. Go to "Logs" tab
  4. View real-time logs and filter by severity

Local development logs are shown in the terminal when running 'gforge dev'.`,
  RunE: func(cmd *cobra.Command, args []string) error {
    banner()
    fmt.Println("Application Logs")
    fmt.Println("────────────────────────────────────────")
    fmt.Println("\n📊 Production Logs (Leapcell):")
    fmt.Println("  1. Visit: https://leapcell.io")
    fmt.Println("  2. Select your app")
    fmt.Println("  3. Go to \"Logs\" tab")
    fmt.Println("  4. View real-time logs and metrics")
    fmt.Println("\n🔧 Local Development Logs:")
    fmt.Println("  Run: gforge dev")
    fmt.Println("  Logs will stream to your terminal")
    fmt.Println("\n💡 Tip: Add structured logging to your app:")
    fmt.Println("  Use slog for better log management in production")
    return nil
  },
}

func init() {
  logsCmd.Flags().BoolVar(&logsFollow, "follow", false, "follow output")
  logsCmd.Flags().StringVar(&logsSince, "since", "", "show logs since timestamp")
  rootCmd.AddCommand(logsCmd)
}
