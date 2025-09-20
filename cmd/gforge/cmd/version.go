package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		if buildDate != "" {
			fmt.Printf("gforge %s (commit %s, built %s)\n", version, commit, buildDate)
		} else {
			fmt.Printf("gforge %s (commit %s)\n", version, commit)
		}
	},
}
