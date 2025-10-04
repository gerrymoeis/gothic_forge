package cmd

import (
  "context"
  "fmt"
  "os"
  "strings"
  "github.com/spf13/cobra"
  "gothicforge3/internal/execx"
)

var (
  testShort bool
  testRace  bool
  testWithBuild bool
)

var testCmd = &cobra.Command{
  Use:   "test",
  Short: "Run tests",
  RunE: func(cmd *cobra.Command, args []string) error {
    banner()
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    if testWithBuild {
      fmt.Println("Running build pre-step...")
      if buildCmd.RunE != nil {
        if err := buildCmd.RunE(buildCmd, nil); err != nil { return err }
      }
    }
    // Silence HTTP request logs during tests for clean output
    _ = os.Setenv("LOG_FORMAT", "off")
    goArgs := []string{"go", "test", "./..."}
    if testShort { goArgs = append(goArgs, "-short") }
    if testRace {
      // Only enable -race when CGO is enabled; otherwise warn and continue without -race
      if out, err := execx.RunCapture(ctx, "go env", "go", "env", "CGO_ENABLED"); err == nil {
        if strings.TrimSpace(out) == "1" {
          goArgs = append(goArgs, "-race")
        } else {
          fmt.Println("Warning: -race requires CGO_ENABLED=1; running tests without -race")
        }
      } else {
        fmt.Println("Warning: unable to detect CGO; running tests without -race")
      }
    }
    return execx.Run(ctx, "go test", goArgs...)
  },
}

func init() {
  testCmd.Flags().BoolVar(&testShort, "short", false, "run short tests")
  testCmd.Flags().BoolVar(&testRace, "race", false, "enable race detector")
  testCmd.Flags().BoolVar(&testWithBuild, "with-build", false, "run build before tests")
  rootCmd.AddCommand(testCmd)
}
