package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	version   = "0.1.0"
	commit    = "dev"
	buildDate = ""
)

// rootCmd is the base command for gforge.
var rootCmd = &cobra.Command{
	Use:   "gforge",
	Short: "Gothic Forge CLI — batteries-included Go web framework",
	Long:  "Gothic Forge CLI — developer-first, secure-by-default framework. Run 'gforge dev' to start hacking.",
}

// Execute runs the CLI.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(devCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(lintCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(releaseCmd)
	rootCmd.AddCommand(addCmd)
}

func banner() {
	color.Cyan("Gothic Forge")
	color.White("All batteries included. Minimal friction. Focus on building.\n")
}
