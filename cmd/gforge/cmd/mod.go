package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"gothicforge3/internal/execx"
)

var modCmd = &cobra.Command{
	Use:   "mod",
	Short: "Go modules helpers",
}

var modTidyCmd = &cobra.Command{
	Use:   "tidy",
	Short: "Run go mod tidy",
	RunE: func(cmd *cobra.Command, args []string) error {
		banner()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		return execx.Run(ctx, "go mod tidy", "go", "mod", "tidy")
	},
}

var modDownloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Run go mod download",
	RunE: func(cmd *cobra.Command, args []string) error {
		banner()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		return execx.Run(ctx, "go mod download", "go", "mod", "download")
	},
}

func init() {
	modCmd.AddCommand(modTidyCmd)
	modCmd.AddCommand(modDownloadCmd)
	rootCmd.AddCommand(modCmd)
}
