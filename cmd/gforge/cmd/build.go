package cmd

import (
	"runtime"

	"gothicforge/internal/framework/execx"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build production server binary",
	RunE: func(cmd *cobra.Command, args []string) error {
		out := "./bin/server"
		if runtime.GOOS == "windows" {
			out += ".exe"
		}
		return execx.Run(cmd.Context(), "build server", "go", "build", "-o", out, "./cmd/server")
	},
}
