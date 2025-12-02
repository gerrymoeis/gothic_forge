package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// QuietMode controls whether to suppress non-essential output for CI/CD
	QuietMode bool
)

var rootCmd = &cobra.Command{
	Use:   "gforge",
	Short: "Gothic Forge CLI",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Disable colors when in quiet mode
		if QuietMode {
			os.Setenv("NO_COLOR", "1")
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&QuietMode, "quiet", "q", false, "minimal output for CI/CD (disables colors, spinners, and progress bars)")
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func banner() {
	if !QuietMode {
		fmt.Println("Gothic Forge v3 :: CLI")
	}
}
