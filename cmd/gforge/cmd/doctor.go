package cmd

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"

    "github.com/fatih/color"
    "github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
    Use:   "doctor",
    Short: "Check your environment for Gothic Forge development",
    RunE: func(cmd *cobra.Command, args []string) error {
        banner()
        color.Cyan("Environment checks:")
        check("Go toolchain", "go")
        check("templ (templates)", "templ")
        check("air (hot reload)", "air")
        check("gotailwindcss (Go Tailwind CLI)", "gotailwindcss")
        check("govulncheck (security)", "govulncheck")
        check("vegeta (load test)", "vegeta")

        color.Cyan("System:")
        fmt.Printf("OS: %s  ARCH: %s\n", runtime.GOOS, runtime.GOARCH)
        return nil
    },
}

func check(title, name string) {
    if _, err := exec.LookPath(name); err == nil {
        color.Green("✔ %s found (%s)", title, name)
        return
    }
    // Fallback: check Go bin dir even if not on PATH
    if binDir := goBinDir(); binDir != "" {
        candidate := filepath.Join(binDir, exeName(name))
        if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
            color.Yellow("⚠ %s found in Go bin but not on PATH (%s)", title, candidate)
            return
        }
    }
    color.Yellow("⚠ %s not found (%s)", title, name)
}
