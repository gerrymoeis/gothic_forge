package cmd

import (
  "fmt"
  "github.com/spf13/cobra"
)

var (
  dbMigrate bool
  dbReset   bool
)

var dbCmd = &cobra.Command{
  Use:   "db",
  Short: "Database helpers (Neon)",
  RunE: func(cmd *cobra.Command, args []string) error {
    banner()
    if dbMigrate {
      fmt.Println("DB: running migrations (placeholder)")
    }
    if dbReset {
      fmt.Println("DB: resetting database (placeholder)")
    }
    if !dbMigrate && !dbReset {
      fmt.Println("Usage: gforge db --migrate | --reset")
    }
    return nil
  },
}

func init() {
  dbCmd.Flags().BoolVar(&dbMigrate, "migrate", false, "apply migrations")
  dbCmd.Flags().BoolVar(&dbReset, "reset", false, "reset database")
  rootCmd.AddCommand(dbCmd)
}
