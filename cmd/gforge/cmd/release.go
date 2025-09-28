package cmd

import (
	"github.com/spf13/cobra"
	"gothicforge/internal/execx"
)

var releaseCmd = &cobra.Command{
	Use:    "release",
	Short:  "Run GoReleaser to build and publish binaries (local or CI)",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return execx.Run(cmd.Context(), "goreleaser", "goreleaser", "release", "--clean")
	},
}
