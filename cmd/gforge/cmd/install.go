package cmd

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "time"

    "gothicforge3/internal/execx"

    "github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
    Use:   "install",
    Short: "Bootstrap the project: deps, tools, styles, env, and optional git init",
    Long:  "Install project dependencies, ensure required tools, scaffold styles and static assets, create .env files, and optionally initialize git.",
    RunE: func(cmd *cobra.Command, args []string) error {
        banner()
        fmt.Println("Install")
        fmt.Println("  • Ensuring Go modules...")
        if err := execx.Run(context.Background(), "go mod tidy", "go", "mod", "tidy"); err != nil {
            return fmt.Errorf("go mod tidy failed: %w", err)
        }

        if os.Getenv("GFORGE_SKIP_TOOLS") == "" && !installSkipTools {
            fmt.Println("  • Installing tools: templ, air, gotailwindcss")
            if _, err := ensureTool("templ", "github.com/a-h/templ/cmd/templ@latest"); err != nil { return err }
            if _, err := ensureTool("air", "github.com/air-verse/air@latest"); err != nil { return err }
            if _, err := ensureTool("gotailwindcss", "github.com/gotailwindcss/tailwind/cmd/gotailwindcss@latest"); err != nil { return err }
        } else {
            fmt.Println("  • Skipping tool installation (GFORGE_SKIP_TOOLS or --skip-tools)")
        }

        fmt.Println("  • Scaffolding styles")
        inputCSS := filepath.Join("app", "styles", "tailwind.input.css")
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

        fmt.Println("  • Creating .env.example and .env if missing")
        envExamplePath := filepath.Join(".env.example")
        if _, err := os.Stat(envExamplePath); os.IsNotExist(err) {
            if err := writeEnvExample(envExamplePath); err != nil { return err }
        }
        envPath := filepath.Join(".env")
        if _, err := os.Stat(envPath); os.IsNotExist(err) {
            // minimal default
            const minimal = "APP_ENV=development\nSITE_BASE_URL=http://127.0.0.1:8080\n"
            if err := os.WriteFile(envPath, []byte(minimal), 0o644); err != nil { return err }
        }
        // Ensure JWT_SECRET exists in .env
        if b, err := os.ReadFile(envPath); err == nil {
            if !strings.Contains(string(b), "JWT_SECRET=") {
                sec := genHex(32)
                f, err := os.OpenFile(envPath, os.O_APPEND|os.O_WRONLY, 0o644)
                if err == nil {
                    defer f.Close()
                    _, _ = f.WriteString("JWT_SECRET=" + sec + "\n")
                }
            }
        }

        fmt.Println("  • Creating sitemap registry (app/sitemap/urls.txt) if missing")
        sitemapDir := filepath.Join("app", "sitemap")
        sitemapFile := filepath.Join(sitemapDir, "urls.txt")
        if _, err := os.Stat(sitemapFile); os.IsNotExist(err) {
            if err := os.MkdirAll(sitemapDir, 0o755); err != nil { return err }
            content := "# Add one path or absolute URL per line.\n# Lines starting with # are ignored.\n/\n"
            if err := os.WriteFile(sitemapFile, []byte(content), 0o644); err != nil { return err }
        }

        if os.Getenv("GFORGE_SKIP_TOOLS") == "" && !installSkipTools {
            fmt.Println("  • Running initial build: templ generate & tailwind build")
            if templPath, err := ensureTool("templ", "github.com/a-h/templ/cmd/templ@latest"); err == nil {
                _ = execx.Run(context.Background(), "templ", templPath, "generate", "-include-version=false", "-include-timestamp=false")
            }
            if gwPath, err := ensureTool("gotailwindcss", "github.com/gotailwindcss/tailwind/cmd/gotailwindcss@latest"); err == nil {
                _ = execx.Run(context.Background(), "gotailwindcss build", gwPath, "build", "-o", "./app/styles/output.css", "./app/styles/tailwind.input.css")
            }
        }

        if installGitInit {
            fmt.Println("  • Initializing git repository")
            // best-effort: ignore errors to keep install idempotent
            _ = execx.Run(context.Background(), "git init", "git", "init")
            _ = execx.Run(context.Background(), "git add .", "git", "add", ".")
            _ = execx.Run(context.Background(), "git commit", "git", "commit", "-m", "chore(install): bootstrap project")
        }

        fmt.Println("────────────────────────────────────────")
        fmt.Println("Install complete.")
        return nil
    },
}

var (
    installSkipTools bool
    installGitInit   bool
)

func init() {
    installCmd.Flags().BoolVar(&installSkipTools, "skip-tools", false, "skip installing CLI tools (templ, air, gotailwindcss)")
    installCmd.Flags().BoolVar(&installGitInit, "git", false, "initialize a git repository and make initial commit")
    rootCmd.AddCommand(installCmd)
}

// genHex returns a random hex string with n bytes.
func genHex(n int) string {
    b := make([]byte, n)
    if _, err := rand.Read(b); err != nil { return "" }
    return hex.EncodeToString(b)
}
