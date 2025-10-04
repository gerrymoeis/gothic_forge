package cmd

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "time"
    "syscall"

    "gothicforge3/internal/execx"
    "gothicforge3/internal/env"

    "github.com/spf13/cobra"
)

var devCmd = &cobra.Command{
    Use:   "dev",
    Short: "Run dev server with hot reload (templ + Air)",
    RunE: func(cmd *cobra.Command, args []string) error {
        banner()
        _ = env.Load()
        host := env.Get("HTTP_HOST", "127.0.0.1")
        port := env.Get("HTTP_PORT", "8080")
        fmt.Printf("Dev: http://%s:%s\n", host, port)
        fmt.Println("Tools: templ • gotailwindcss")

        ctx, cancel := context.WithCancel(context.Background())
        defer cancel()

        // Ensure Go modules so a fresh clone can just run `gforge dev`
        _ = execx.Run(ctx, "go mod tidy", "go", "mod", "tidy")

		// Handle Ctrl+C
		go func() {
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
			<-ch
			fmt.Println("\nshutting down...")
			cancel()
		}()

		        // templ generate once (stable); no watch to avoid dev crashes
        if templPath, err := ensureTool("templ", "github.com/a-h/templ/cmd/templ@latest"); err == nil {
            _ = execx.Run(ctx, "templ", templPath, "generate", "-include-version=false", "-include-timestamp=false")
        } else {
            fmt.Printf("templ auto-install failed: %v\n", err)
        }
        // Tailwind CSS build via gotailwindcss (generate app/styles/output.css from app/styles/tailwind.input.css)
        go func() {
            gwPath, err := ensureTool("gotailwindcss", "github.com/gotailwindcss/tailwind/cmd/gotailwindcss@latest")
            if err != nil {
                fmt.Printf("gotailwindcss not available: %v\n", err)
                return
            }
            input := "./app/styles/tailwind.input.css"
            output := "./app/styles/output.css"
            // initial build
            _ = execx.Run(ctx, "gotailwindcss build", gwPath, "build", "-o", output, input)
            var lastMod time.Time
            for {
                select { case <-ctx.Done(): return; default: }
                if fi, err := os.Stat(input); err == nil {
                    if fi.ModTime().After(lastMod) {
                        _ = execx.Run(ctx, "gotailwindcss build", gwPath, "build", "-o", output, input)
                        lastMod = fi.ModTime()
                    }
                }
                time.Sleep(1 * time.Second)
            }
        }()

        // Server: run directly (skip Air to avoid interference with templ)
        go func() {
            fmt.Println("Server: go run")
            _ = execx.Run(ctx, "server", "go", "run", "./cmd/server")
        }()

        fmt.Println("Watching for changes...")

        <-ctx.Done()
        time.Sleep(200 * time.Millisecond)
        return nil
    },
}

func init() { rootCmd.AddCommand(devCmd) }
