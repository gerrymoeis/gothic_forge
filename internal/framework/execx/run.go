package execx

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
)

// Look finds an executable in PATH.
func Look(name string) (string, bool) {
	p, err := exec.LookPath(name)
	return p, err == nil
}

// Run streams a command's stdout/stderr to the console with prefixes.
func Run(ctx context.Context, title string, name string, args ...string) error {
	color.Cyan("==> %s", title)
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = os.Environ()
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		return err
	}

	outDone := make(chan struct{})
	errDone := make(chan struct{})

	go func() {
		s := bufio.NewScanner(stdout)
		for s.Scan() {
			fmt.Println(s.Text())
		}
		close(outDone)
	}()
	go func() {
		s := bufio.NewScanner(stderr)
		for s.Scan() {
			color.HiBlack(s.Text())
		}
		close(errDone)
	}()

	<-outDone
	<-errDone
	return cmd.Wait()
}

// Shell returns the system shell and arg flag to run a command string.
func Shell() (string, string) {
	if runtime.GOOS == "windows" {
		return "cmd", "/C"
	}
	return "bash", "-lc"
}

// TimeoutContext is a helper to build a context with timeout.
func TimeoutContext(d time.Duration) (context.Context, context.CancelFunc) {
    if d <= 0 {
        return context.WithCancel(context.Background())
    }
    return context.WithTimeout(context.Background(), d)
}

// JoinCommand builds a shell command string.
func JoinCommand(parts ...string) string {
	return strings.Join(parts, " ")
}
