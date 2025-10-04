package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func exeName(name string) string {
	if runtime.GOOS == "windows" { return name + ".exe" }
	return name
}

func goEnv(key string) string {
	out, err := exec.Command("go", "env", key).Output()
	if err != nil { return "" }
	return strings.TrimSpace(string(out))
}

func goBinDir() string {
	if bin := goEnv("GOBIN"); bin != "" { return bin }
	gopath := goEnv("GOPATH")
	if gopath == "" { return "" }
	return filepath.Join(gopath, "bin")
}

// ensureTool ensures a CLI tool is available. If missing, it runs `go install <module>`
// and returns the absolute path to the installed binary (from GOBIN or GOPATH/bin).
func ensureTool(name, module string) (string, error) {
	if p, err := exec.LookPath(name); err == nil {
		return p, nil
	}
	fmt.Printf("%s not found. Installing %s...\n", name, module)
	cmd := exec.Command("go", "install", module)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("install failed for %s: %w", name, err)
	}
	if binDir := goBinDir(); binDir != "" {
		bin := filepath.Join(binDir, exeName(name))
		if _, err := os.Stat(bin); err == nil {
			return bin, nil
		}
	}
	if p, err := exec.LookPath(name); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("%s installed but not found in PATH; ensure GOBIN/GOPATH/bin is on PATH", name)
}
