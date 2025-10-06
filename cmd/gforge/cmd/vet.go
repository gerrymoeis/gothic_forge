package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"gothicforge3/internal/execx"
)

var vetCmd = &cobra.Command{
	Use:   "vet",
	Short: "Run go vet ./...",
	RunE: func(cmd *cobra.Command, args []string) error {
		banner()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		return execx.Run(ctx, "go vet", "go", "vet", "./...")
	},
}

func init() { rootCmd.AddCommand(vetCmd) }
