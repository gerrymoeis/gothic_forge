package cmd

import (
	"github.com/spf13/cobra"
	"gothicforge/internal/framework/execx"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run unit tests",
	RunE: func(cmd *cobra.Command, args []string) error {
		return execx.Run(cmd.Context(), "go test", "go", "test", "./...")
	},
}
