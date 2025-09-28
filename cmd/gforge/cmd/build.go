package cmd

import (
	"runtime"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gothicforge/internal/execx"
)

var buildCmd = &cobra.Command{
	Use:    "build",
	Short:  "Build production server binary",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Ensure templ code is generated once before build.
		if err := execx.Run(cmd.Context(), "templ generate", "templ", "generate", "-include-version=false", "-include-timestamp=false"); err != nil {
			color.Yellow("templ not available. Installing and retrying...")
			_ = execx.Run(cmd.Context(), "go install templ", "go", "install", "github.com/a-h/templ/cmd/templ@latest")
			if err2 := execx.Run(cmd.Context(), "templ generate", "templ", "generate", "-include-version=false", "-include-timestamp=false"); err2 != nil {
				return err2
			}
		}
		out := "./bin/server"
		if runtime.GOOS == "windows" {
			out += ".exe"
		}
		return execx.Run(cmd.Context(), "build server", "go", "build", "-o", out, "./cmd/server")
	},
}
