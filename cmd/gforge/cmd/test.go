package cmd

import (
    "bufio"
    "context"
    "encoding/json"
    "fmt"
    "os"
    "os/exec"
    "strings"
    "time"

    "github.com/fatih/color"
    "github.com/spf13/cobra"
    "gothicforge/internal/execx"
)

var testCmd = &cobra.Command{
    Use:   "test",
    Short: "Run unit tests (pretty logger)",
    RunE: func(cmd *cobra.Command, args []string) error {
        return runGoTests(cmd.Context())
    },
}

type testEvent struct {
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Elapsed float64 `json:"Elapsed"`
	Output  string  `json:"Output"`
}

func runGoTests(ctx context.Context) error {
    // Ensure templ code is generated once before running tests
    if templPath, err := ensureTool("templ", "github.com/a-h/templ/cmd/templ@latest"); err == nil {
        _ = execx.Run(ctx, "templ generate", templPath, "generate", "-include-version=false", "-include-timestamp=false")
    } else {
        color.Yellow("templ not available and auto-install failed: %v", err)
    }

    // Discover only packages that actually contain tests to avoid noisy
    // "[no test files]" lines. Fallback to ./... if discovery fails or none found.
    listCmd := exec.CommandContext(ctx, "go", "list", "-f", "{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}", "./...")
    listCmd.Env = os.Environ()
    out, _ := listCmd.Output()
    var testPkgs []string
    for _, line := range strings.Split(string(out), "\n") {
        line = strings.TrimSpace(line)
        if line != "" {
            testPkgs = append(testPkgs, line)
        }
    }
    if len(testPkgs) == 0 {
        testPkgs = []string{"./..."}
    }

    color.Cyan("==> Running tests on %d package(s)", len(testPkgs))
    start := time.Now()
    args := append([]string{"test", "-json"}, testPkgs...)
    cmd := exec.CommandContext(ctx, "go", args...)
    cmd.Env = os.Environ()
    stdout, _ := cmd.StdoutPipe()
    stderr, _ := cmd.StderrPipe()
    if err := cmd.Start(); err != nil {
        return err
    }

	// Stats
	pkgs := map[string]struct{}{}
	var pass, fail, skip, total int
	outDone := make(chan error, 1)
	go func() {
		s := bufio.NewScanner(stdout)
		for s.Scan() {
			line := s.Text()
			var ev testEvent
			if err := json.Unmarshal([]byte(line), &ev); err != nil {
				// Non-JSON output; print dim
				color.HiBlack(line)
				continue
			}
			if ev.Package != "" {
				pkgs[ev.Package] = struct{}{}
			}
			if ev.Test != "" {
				switch ev.Action {
				case "run":
					// show test name in dim
					color.HiBlack("• %s", ev.Test)
				case "pass":
					pass++
					total++
					color.Green("  PASS %s (%.2fs)", ev.Test, ev.Elapsed)
				case "fail":
					fail++
					total++
					color.Red("  FAIL %s (%.2fs)", ev.Test, ev.Elapsed)
				case "skip":
					skip++
					total++
					color.Yellow("  SKIP %s", ev.Test)
				}
			}
			if ev.Output != "" && strings.HasPrefix(ev.Output, "--- ") {
				color.HiBlack(strings.TrimSpace(ev.Output))
			}
		}
		outDone <- s.Err()
	}()

	// Read stderr (tooling info)
	errDone := make(chan error, 1)
	go func() {
		s := bufio.NewScanner(stderr)
		for s.Scan() {
			color.HiBlack(s.Text())
		}
		errDone <- s.Err()
	}()

	// Wait for completion
	_ = <-outDone
	_ = <-errDone
	if err := cmd.Wait(); err != nil {
		// Even when some tests fail, we still want to print the summary
		printSummary(pkgs, pass, fail, skip, time.Since(start))
		return err
	}
	printSummary(pkgs, pass, fail, skip, time.Since(start))
	if fail > 0 {
		return fmt.Errorf("tests failed: %d", fail)
	}
	return nil
}

func printSummary(pkgs map[string]struct{}, pass, fail, skip int, dur time.Duration) {
	total := pass + fail + skip
	color.Cyan("\n==> Test Summary")
	fmt.Printf("Packages: %d\n", len(pkgs))
	fmt.Printf("Total: %d  Passed: %d  Failed: %d  Skipped: %d  Duration: %s\n",
		total, pass, fail, skip, dur.Round(time.Millisecond))
}
