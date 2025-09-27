package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var addDBCmd = &cobra.Command{
	Use:   "db",
	Short: "Scaffold database migrations folder (pure SQL)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := filepath.Join("app", "db", "migrations")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		// Create an initial migration only if the directory is empty
		entries, err := os.ReadDir(dir)
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			path := filepath.Join(dir, "0001_init.sql")
			content := `-- Gothic Forge initial migration (example)
-- Write your SQL here. Keep migrations pure SQL for portability and clarity.
-- Example:
-- CREATE TABLE example (
--   id SERIAL PRIMARY KEY,
--   name TEXT NOT NULL
-- );
`
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				return err
			}
			color.Green("✔ Created %s", path)
		} else {
			color.Yellow("migrations directory already has files; no initial migration created")
		}
		fmt.Println("Migrations directory:", dir)
		return nil
	},
}

func init() {
	addCmd.AddCommand(addDBCmd)
}
