package cmd

import (
	"context"
	"fmt"
	"strings"
	"github.com/spf13/cobra"
	"gothicforge3/internal/execx"
)

var (
	lintArgs string
)

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Run golangci-lint (auto-installs if missing)",
	RunE: func(cmd *cobra.Command, args []string) error {
		banner()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		// Ensure golangci-lint
		if _, err := ensureTool("golangci-lint", "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"); err != nil {
			return fmt.Errorf("ensure golangci-lint: %w", err)
		}
		// Run
		if lintArgs == "" {
			return execx.Run(ctx, "golangci-lint run", "golangci-lint", "run")
		}
		extra := strings.Fields(lintArgs)
		return execx.Run(ctx, "golangci-lint run", append([]string{"golangci-lint", "run"}, extra...)...)
	},
}

func init() {
	lintCmd.Flags().StringVar(&lintArgs, "args", "", "extra args to pass to golangci-lint run")
	rootCmd.AddCommand(lintCmd)
}
