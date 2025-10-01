package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gothicforge/internal/execx"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Tool metadata for listing/installation.
type toolMeta struct {
	Name        string
	Description string
	IsGo        bool
	Module      string // go install path when IsGo
	Manual      string // manual install URL/instruction when not IsGo
}

var (
	goTools = []toolMeta{
		{Name: "templ", Description: "Templ code generator (templates)", IsGo: true, Module: "github.com/a-h/templ/cmd/templ@latest"},
		{Name: "air", Description: "Hot reload (dev server)", IsGo: true, Module: "github.com/air-verse/air@latest"},
		{Name: "gotailwindcss", Description: "Pure-Go Tailwind CSS builder", IsGo: true, Module: "github.com/gotailwindcss/tailwind/cmd/gotailwindcss@latest"},
		{Name: "render", Description: "Render CLI (deploy provider)", IsGo: true, Module: "", Manual: "https://render.com/docs/cli"},
		{Name: "govulncheck", Description: "Go vulnerability scanner", IsGo: true, Module: "golang.org/x/vuln/cmd/govulncheck@latest"},
		{Name: "vegeta", Description: "HTTP load testing tool", IsGo: true, Module: "github.com/tsenart/vegeta/v12@latest"},
	}
	manualTools = []toolMeta{
		{Name: "ffmpeg", Description: "Media optimizer (video/audio)", Manual: "https://ffmpeg.org/download.html"},
		{Name: "jpegoptim", Description: "JPEG optimizer (strip/progressive)", Manual: "https://github.com/tjko/jpegoptim"},
		{Name: "oxipng", Description: "PNG optimizer (lossless)", Manual: "https://github.com/shssoichiro/oxipng"},
		{Name: "gifsicle", Description: "GIF optimizer", Manual: "https://www.lcdf.org/gifsicle/"},
	}
)

var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "Manage external tools (list/install)",
}

var toolsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List required/optional external tools and their status",
	RunE: func(cmd *cobra.Command, args []string) error {
		banner()
		color.Cyan("Core & recommended tools:")
		printTools(goTools)
		fmt.Println()
		color.Cyan("Optional media optimizers (used by 'gforge export'):")
		printTools(manualTools)
		return nil
	},
}

var toolsInstallCmd = &cobra.Command{
	Use:   "install [name|all|deploy]",
	Short: "Install a Go-based tool, the deploy stack (Render), or all",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		banner()
		if len(args) == 0 {
			return errors.New("missing argument: specify a tool name or 'all'")
		}
		target := strings.ToLower(strings.TrimSpace(args[0]))
		if target == "all" {
			// Install all Go-based tools
			for _, t := range goTools {
				if t.IsGo {
					if _, ok := execx.Look(t.Name); ok {
						color.Green("✔ %s already installed", t.Name)
						continue
					}
					// Special-case render (no go-installable module path)
					if strings.EqualFold(t.Name, "render") {
						if err := installRenderFromSource(cmd); err != nil {
							color.Red("✖ failed to install %s: %v", t.Name, err)
							return err
						}
						color.Green("✔ installed %s", t.Name)
						continue
					}
					if t.Module != "" {
						if _, err := ensureTool(t.Name, t.Module); err != nil {
							color.Red("✖ failed to install %s: %v", t.Name, err)
							return err
						}
						color.Green("✔ installed %s", t.Name)
					}
				}
			}
			return nil
		}
		if target == "deploy" {
			// Omakase deploy stack: ensure Render CLI
			if _, ok := execx.Look("render"); !ok {
				// Try build-from-source fallback
				if err := installRenderFromSource(cmd); err != nil {
					color.Red("✖ failed to install render CLI: %v", err)
					color.Yellow("Install manually from releases or docs: https://render.com/docs/cli")
					return err
				}
			}
			color.Green("✔ render CLI is available")
			color.HiGreen("Deploy stack ready (Render CLI). Next steps:")
			color.HiBlack("1) Run 'render login' to authenticate (browser flow)")
			color.HiBlack("2) Create a New Web Service: https://dashboard.render.com/web/new?newUser=true")
			color.HiBlack("3) Set env: FIBER_HOST=0.0.0.0, FIBER_PORT=$PORT, and your BASE_URL/DATABASE_URL/REDIS_URL if any")
			color.HiBlack("4) Build: go run github.com/a-h/templ/cmd/templ@latest generate -include-version=false -include-timestamp=false && go build -o server ./cmd/server")
			color.HiBlack("   Start: ./server")
			if isInteractive() {
				fmt.Print("Open Render New Web Service page now? [Y/n]: ")
				var ans string
				fmt.Scanln(&ans)
				ans = strings.TrimSpace(strings.ToLower(ans))
				if ans == "" || ans == "y" || ans == "yes" {
					_ = openURL("https://dashboard.render.com/web/new?newUser=true", cmd)
				}
			}
			return nil
		}
		// Install a single tool by name
		if tm, ok := findTool(target); ok {
			if tm.IsGo {
				if _, ok := execx.Look(tm.Name); ok {
					color.Green("✔ %s already installed", tm.Name)
					return nil
				}
				// Special-case render: build from source
				if strings.EqualFold(tm.Name, "render") {
					if err := installRenderFromSource(cmd); err != nil {
						color.Red("✖ failed to install %s: %v", tm.Name, err)
						return err
					}
					color.Green("✔ installed %s", tm.Name)
					return nil
				}
				// Default path for other Go tools
				if tm.Module != "" {
					if _, err := ensureTool(tm.Name, tm.Module); err != nil {
						color.Red("✖ failed to install %s: %v", tm.Name, err)
						return err
					}
					color.Green("✔ installed %s", tm.Name)
					return nil
				}
			}
			color.Yellow("%s is not Go-installable. Install manually: %s", tm.Name, tm.Manual)
			return nil
		}
		return fmt.Errorf("unknown tool: %s", target)
	},
}

func findTool(name string) (toolMeta, bool) {
	for _, t := range goTools {
		if strings.EqualFold(t.Name, name) {
			return t, true
		}
	}
	for _, t := range manualTools {
		if strings.EqualFold(t.Name, name) {
			return t, true
		}
	}
	return toolMeta{}, false
}

func printTools(list []toolMeta) {
	// stable order
	items := make([]toolMeta, len(list))
	copy(items, list)
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	for _, t := range items {
		if _, ok := execx.Look(t.Name); ok {
			color.Green("✔ %s — %s", t.Name, t.Description)
			continue
		}
		if t.IsGo {
			if t.Module != "" {
				color.Yellow("⚠ %s — %s (go install: %s)", t.Name, t.Description, t.Module)
			} else {
				color.Yellow("⚠ %s — %s (build from source: %s)", t.Name, t.Description, t.Manual)
			}
		} else {
			color.Yellow("⚠ %s — %s (manual: %s)", t.Name, t.Description, t.Manual)
		}
	}
}

// installRenderFromSource clones the Render CLI and builds the binary into the user's Go bin dir.
func installRenderFromSource(cmd *cobra.Command) error {
    // Preconditions
    if _, ok := execx.Look("git"); !ok {
        return errors.New("git not found in PATH; install Git to build Render CLI from source")
    }
    binDir := goBinDir()
    if binDir == "" {
        return errors.New("unable to determine Go bin dir (GOBIN or GOPATH/bin); set one and ensure it's on PATH")
    }

    // Temp workspace
    tmp, err := os.MkdirTemp("", "rendercli-*")
    if err != nil { return err }
    defer os.RemoveAll(tmp)

    repoDir := filepath.Join(tmp, "cli")
    if err := execx.Run(cmd.Context(), "git clone render-oss/cli", "git", "clone", "--depth", "1", "https://github.com/render-oss/cli.git", repoDir); err != nil {
        return err
    }

    // Build the CLI (use go -C for cross-platform path handling)
    outPath := filepath.Join(tmp, exeName("render"))
    if err := execx.Run(cmd.Context(), "go build render CLI", "go", "build", "-C", repoDir, "-o", outPath, "."); err != nil {
        return err
    }

    // Move to bin dir (copy to handle cross-volume)
    dest := filepath.Join(binDir, exeName("render"))
    if err := copyFile(outPath, dest); err != nil {
        return err
    }
    color.Green("✔ built render -> %s", dest)
    color.HiBlack("Ensure %s is in your PATH", filepath.Dir(dest))
    return nil
}

func init() {
	toolsCmd.AddCommand(toolsListCmd)
	toolsCmd.AddCommand(toolsInstallCmd)
	rootCmd.AddCommand(toolsCmd)
}
