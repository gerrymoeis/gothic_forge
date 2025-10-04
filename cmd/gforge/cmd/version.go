package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// These can be set via -ldflags at build time
	Version   = "dev"
	Commit    = ""
	BuildDate = ""
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print gforge version",
	Run: func(cmd *cobra.Command, args []string) {
		banner()
		fmt.Println("Version")
		fmt.Printf("  • gforge:   %s\n", Version)
		if Commit != "" { fmt.Printf("  • commit:   %s\n", Commit) }
		if BuildDate != "" { fmt.Printf("  • built:    %s\n", BuildDate) }
		fmt.Println("────────────────────────────────────────")
	},
}

func init() { rootCmd.AddCommand(versionCmd) }
