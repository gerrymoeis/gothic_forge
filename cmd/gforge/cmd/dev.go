package cmd

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "path/filepath"
    "strings"
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

		        // templ generate once
        if templPath, err := ensureTool("templ", "github.com/a-h/templ/cmd/templ@latest"); err == nil {
            _ = execx.Run(ctx, "templ", templPath, "generate", "-include-version=false", "-include-timestamp=false")
        } else { fmt.Printf("templ auto-install failed: %v\n", err) }
        // Watch for .templ changes and re-generate
        go func() {
            last := map[string]time.Time{}
            for {
                select { case <-ctx.Done(): return; default: }
                entries, err := os.ReadDir(filepath.Join("app", "templates"))
                if err == nil {
                    changed := false
                    for _, e := range entries {
                        if e.IsDir() { continue }
                        name := e.Name()
                        if !strings.HasSuffix(name, ".templ") { continue }
                        p := filepath.Join("app", "templates", name)
                        if fi, err := os.Stat(p); err == nil {
                            mt := fi.ModTime()
                            if prev, ok := last[p]; !ok || mt.After(prev) {
                                last[p] = mt
                                changed = true
                            }
                        }
                    }
                    if changed {
                        if templPath, err := ensureTool("templ", "github.com/a-h/templ/cmd/templ@latest"); err == nil {
                            _ = execx.Run(ctx, "templ", templPath, "generate", "-include-version=false", "-include-timestamp=false")
                        }
                    }
                }
                time.Sleep(1 * time.Second)
            }
        }()
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

        if devAir {
            // Use Air for full autoreload
            if airPath, err := ensureTool("air", "github.com/air-verse/air@latest"); err == nil {
                fmt.Println("Server: air")
                go func() { _ = execx.Run(ctx, "air", airPath, "-c", ".air.toml") }()
            } else {
                fmt.Printf("air not available: %v\nfalling back to go run\n", err)
                go func() { fmt.Println("Server: go run"); _ = execx.Run(ctx, "server", "go", "run", "./cmd/server") }()
            }
        } else {
            // Run go server and restart on component changes
            srvCtx, srvCancel := context.WithCancel(ctx)
            startServer := func() { go func() { fmt.Println("Server: go run"); _ = execx.Run(srvCtx, "server", "go", "run", "./cmd/server") }() }
            startServer()
            go func() {
                last := map[string]time.Time{}
                lastRestart := time.Now()
                scan := func(dir string) {
                    entries, err := os.ReadDir(dir)
                    if err != nil { return }
                    for _, e := range entries {
                        if e.IsDir() { continue }
                        name := e.Name()
                        if !strings.HasSuffix(name, ".go") { continue }
                        p := filepath.Join(dir, name)
                        if fi, err := os.Stat(p); err == nil {
                            mt := fi.ModTime()
                            if prev, ok := last[p]; !ok || mt.After(prev) { last[p] = mt }
                        }
                    }
                }
                // Prime snapshot
                scan(filepath.Join("app", "templates"))
                scan(filepath.Join("app", "routes"))
                for {
                    select { case <-ctx.Done(): return; default: }
                    changed := false
                    // rescan and detect any file newer than snapshot
                    for _, dir := range []string{filepath.Join("app", "templates"), filepath.Join("app", "routes")} {
                        entries, err := os.ReadDir(dir)
                        if err != nil { continue }
                        for _, e := range entries {
                            if e.IsDir() { continue }
                            name := e.Name(); if !strings.HasSuffix(name, ".go") { continue }
                            p := filepath.Join(dir, name)
                            if fi, err := os.Stat(p); err == nil {
                                mt := fi.ModTime(); prev := last[p]
                                if mt.After(prev) { last[p] = mt; changed = true }
                            }
                        }
                    }
                    if changed && time.Since(lastRestart) > 500*time.Millisecond {
                        fmt.Println("Restarting server...")
                        srvCancel()
                        srvCtx, srvCancel = context.WithCancel(ctx)
                        startServer()
                        lastRestart = time.Now()
                    }
                    time.Sleep(1 * time.Second)
                }
            }()
        }

        fmt.Println("Watching for changes...")

        <-ctx.Done()
        time.Sleep(200 * time.Millisecond)
        return nil
    },
}

var devAir bool

func init() {
    devCmd.Flags().BoolVar(&devAir, "air", false, "Use Air for full autoreload (requires .air.toml)")
    rootCmd.AddCommand(devCmd)
}
