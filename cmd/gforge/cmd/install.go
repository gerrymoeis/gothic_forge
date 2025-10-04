package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gothicforge3/internal/execx"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install tools and bootstrap app assets (deprecated; use 'gforge doctor --fix')",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		banner()
		fmt.Println("Setup")
		fmt.Println("  • Ensuring Go modules...")
		if err := execx.Run(context.Background(), "go mod tidy", "go", "mod", "tidy"); err != nil {
			return fmt.Errorf("go mod tidy failed: %w", err)
		}

		fmt.Println("  • Installing tools: templ, air, gotailwindcss")
		if _, err := ensureTool("templ", "github.com/a-h/templ/cmd/templ@latest"); err != nil {
			return err
		}
		if _, err := ensureTool("air", "github.com/air-verse/air@latest"); err != nil {
			return err
		}
		if _, err := ensureTool("gotailwindcss", "github.com/gotailwindcss/tailwind/cmd/gotailwindcss@latest"); err != nil {
			return err
		}

		fmt.Println("  • Scaffolding styles")
		inputCSS := filepath.Join("app", "styles", "input.css")
		css := "@import \"tailwindcss\" source(none);\n@source \"./app/**/*.{templ,go,html}\";\n@plugin \"./daisyui.js\";\n"
		if err := execx.WriteFileIfMissing(inputCSS, []byte(css), 0o644); err != nil {
			return err
		}

		fmt.Println("  • Fetching daisyUI plugin (optional)")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		daisy := filepath.Join("app", "styles", "daisyui.js")
		if _, err := os.Stat(daisy); os.IsNotExist(err) {
			_ = execx.Download(ctx, "https://github.com/saadeghi/daisyui/releases/latest/download/daisyui.js", daisy)
		}
		daisyTheme := filepath.Join("app", "styles", "daisyui-theme.js")
		if _, err := os.Stat(daisyTheme); os.IsNotExist(err) {
			_ = execx.Download(ctx, "https://github.com/saadeghi/daisyui/releases/latest/download/daisyui-theme.js", daisyTheme)
		}

		fmt.Println("  • Creating static assets")
		fav := filepath.Join("app", "static", "favicon.svg")
		favicon := "<svg xmlns=\"http://www.w3.org/2000/svg\" viewBox=\"0 0 24 24\"><circle cx=\"12\" cy=\"12\" r=\"10\" fill=\"#4b5563\"/><text x=\"12\" y=\"16\" text-anchor=\"middle\" font-size=\"10\" fill=\"white\">GF</text></svg>\n"
		if err := execx.WriteFileIfMissing(fav, []byte(favicon), 0o644); err != nil {
			return err
		}

		fmt.Println("────────────────────────────────────────")
		fmt.Println("Install complete.")
		return nil
	},
}

func init() { rootCmd.AddCommand(installCmd) }
