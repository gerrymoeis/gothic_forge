package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [feature]",
	Short: "Scaffold optional features (e.g., redis, auth, db)",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			fmt.Println("Available features: redis, auth, db")
			fmt.Println("Subcommands:")
			fmt.Println("  gforge add page <name>   # Scaffold a new Templ page, wire route, and register for SSG")
			return nil
		}
		switch args[0] {
		case "redis":
			fmt.Println("TODO: scaffold Redis integration (pool, health check, cache helpers)")
		case "auth":
			fmt.Println("TODO: scaffold auth routes and session helpers")
		case "db":
			fmt.Println("TODO: scaffold database migrations and repository examples")
		default:
			fmt.Printf("Unknown feature: %s\n", args[0])
		}
		return nil
	},
}
