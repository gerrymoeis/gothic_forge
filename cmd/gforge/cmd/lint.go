package cmd

import (
	"github.com/spf13/cobra"
	"gothicforge/internal/framework/execx"
)

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Run basic linters (go vet, fmt check)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := execx.Run(cmd.Context(), "go vet", "go", "vet", "./..."); err != nil {
			return err
		}
		return execx.Run(cmd.Context(), "gofmt check", "gofmt", "-l", ".")
	},
}
