package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var addAPICmd = &cobra.Command{
	Use:   "api <name>",
	Short: "Scaffold a minimal JSON API route registrant (GET /api/<name>)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		raw := args[0]
		name := sanitizeKebab(raw)
		if name == "" {
			return fmt.Errorf("invalid api name: %q", raw)
		}
		comp := toPascal(name)

		rDir := filepath.Join("app", "routes")
		if err := os.MkdirAll(rDir, 0o755); err != nil {
			return err
		}
		rPath := filepath.Join(rDir, name+"_api.go")
		if _, err := os.Stat(rPath); err == nil {
			color.Yellow("API registrant already exists: %s", rPath)
			return nil
		}
		code := fmt.Sprintf(`package routes

import (
    "github.com/gofiber/fiber/v2"
)

func init() {
    RegisterRoute(func(app *fiber.App) {
        app.Get("/api/%s", func(c *fiber.Ctx) error {
            return c.JSON(fiber.Map{
                "ok": true,
                "data": fiber.Map{
                    "message": "%s endpoint",
                },
            })
        })
    })
}
`, "%s", "%s")
		code = fmt.Sprintf(code, name, comp)
		if err := os.WriteFile(rPath, []byte(code), 0o644); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	addCmd.AddCommand(addAPICmd)
}
