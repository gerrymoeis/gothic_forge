package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"gothicforge3/internal/execx"
)

var vulnCmd = &cobra.Command{
	Use:   "vuln",
	Short: "Run govulncheck ./... (auto-installs if missing)",
	RunE: func(cmd *cobra.Command, args []string) error {
		banner()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if _, err := ensureTool("govulncheck", "golang.org/x/vuln/cmd/govulncheck@latest"); err != nil {
			return fmt.Errorf("ensure govulncheck: %w", err)
		}
		return execx.Run(ctx, "govulncheck", "govulncheck", "./...")
	},
}

func init() { rootCmd.AddCommand(vulnCmd) }
