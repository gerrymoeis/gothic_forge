package cmd

import (
	"github.com/spf13/cobra"
	"gothicforge/internal/framework/execx"
)

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Run GoReleaser to build and publish binaries (local or CI)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return execx.Run(cmd.Context(), "goreleaser", "goreleaser", "release", "--clean")
	},
}
