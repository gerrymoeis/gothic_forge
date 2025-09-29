package cmd

import (
	"fmt"
	"runtime"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gothicforge/internal/execx"
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
		color.Cyan("Required for deploy:")
		printTool("flyctl")
		color.Cyan("Optional media optimizers (used by export):")
		for _, t := range []string{"ffmpeg", "jpegoptim", "oxipng", "gifsicle"} {
			printTool(t)
		}
		return nil
	},
}

var toolsInstallYes bool

var toolsInstallCmd = &cobra.Command{
	Use:   "install [tool|group]...",
	Short: "Install tools via your OS package manager (groups: media, deploy, all)",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		banner()
		if len(args) == 0 {
			color.Yellow("No tools specified. Examples:\n  gforge tools install media\n  gforge tools install flyctl\n  gforge tools install ffmpeg oxipng")
			return nil
		}

		// Expand groups
		expanded := expandGroups(args)
		// Dedupe and sort for stability
		uniq := dedupe(expanded)
		sort.Strings(uniq)

		// Build plan
		plan := buildInstallPlan(uniq)
		if len(plan.Commands) == 0 && len(plan.Manual) == 0 {
			color.Green("All requested tools are already installed.")
			return nil
		}

		color.Cyan("Planned installations (OS=%s):", runtime.GOOS)
		for _, c := range plan.Commands {
			fmt.Printf("  $ %s %s\n", c.Name, strings.Join(c.Args, " "))
		}
		if len(plan.Manual) > 0 {
			color.Yellow("Manual steps (no supported package manager found for these tools):")
			for _, m := range plan.Manual {
				fmt.Printf("  - %s\n", m)
			}
		}

		if !toolsInstallYes {
			color.Yellow("Dry run only. Re-run with --yes to execute.")
			return nil
		}

		// Execute commands sequentially
		for _, c := range plan.Commands {
			title := fmt.Sprintf("install %s", strings.Join(uniq, ", "))
			if err := execx.Run(cmd.Context(), title, c.Name, c.Args...); err != nil {
				return err
			}
		}

		color.HiGreen("Completed.")
		return nil
	},
}

func init() {
	toolsInstallCmd.Flags().BoolVar(&toolsInstallYes, "yes", false, "execute without prompt (non-interactive)")

	toolsCmd.AddCommand(toolsListCmd)
	toolsCmd.AddCommand(toolsInstallCmd)
	rootCmd.AddCommand(toolsCmd)
}

// --- utilities ---

func printTool(name string) {
	if has(name) {
		color.Green("✔ %s found", name)
	} else {
		color.Yellow("⚠ %s not found", name)
	}
}

func has(name string) bool { _, ok := execx.Look(name); return ok }

type installCommand struct {
	Name string
	Args []string
}

type installPlan struct {
	Commands []installCommand // commands to run
	Manual   []string         // manual instructions if no package manager
}

func expandGroups(args []string) []string {
	var out []string
	for _, a := range args {
		switch strings.ToLower(strings.TrimSpace(a)) {
		case "media":
			out = append(out, "ffmpeg", "jpegoptim", "oxipng", "gifsicle")
		case "deploy":
			out = append(out, "flyctl")
		case "all":
			out = append(out, "ffmpeg", "jpegoptim", "oxipng", "gifsicle", "flyctl")
		default:
			out = append(out, a)
		}
	}
	return out
}

func dedupe(in []string) []string {
	m := map[string]struct{}{}
	var out []string
	for _, v := range in {
		key := strings.ToLower(strings.TrimSpace(v))
		if key == "" {
			continue
		}
		if _, ok := m[key]; ok {
			continue
		}
		m[key] = struct{}{}
		out = append(out, key)
	}
	return out
}

func buildInstallPlan(tools []string) installPlan {
	var plan installPlan
	// Skip already installed
	var needed []string
	for _, t := range tools {
		if has(t) {
			continue
		}
		needed = append(needed, t)
	}
	if len(needed) == 0 {
		return plan
	}

	os := runtime.GOOS
	switch os {
	case "windows":
		// Try winget, then choco, then scoop per-tool so we don't fail a tool just because the first manager lacks it.
		hasWinget := has("winget")
		hasChoco := has("choco")
		hasScoop := has("scoop")
		for _, t := range needed {
			added := false
			if hasWinget {
				if cmd, ok := wingetCmd(t); ok {
					plan.Commands = append(plan.Commands, cmd)
					added = true
				}
			}
			if !added && hasChoco {
				if cmd, ok := chocoCmd(t); ok {
					plan.Commands = append(plan.Commands, cmd)
					added = true
				}
			}
			if !added && hasScoop {
				if cmd, ok := scoopCmd(t); ok {
					plan.Commands = append(plan.Commands, cmd)
					added = true
				}
			}
			if !added {
				plan.Manual = append(plan.Manual, manualURL(t))
			}
		}

	case "darwin":
		if has("brew") {
			for _, t := range needed {
				if cmd, ok := brewCmd(t); ok {
					plan.Commands = append(plan.Commands, cmd)
				} else {
					plan.Manual = append(plan.Manual, manualURL(t))
				}
			}
		} else {
			for _, t := range needed {
				plan.Manual = append(plan.Manual, manualURL(t))
			}
		}

	case "linux":
		// Try apt, then dnf, then pacman
		if has("apt-get") {
			// Add update once at top
			plan.Commands = append(plan.Commands, installCommand{Name: "sudo", Args: []string{"apt-get", "update"}})
			for _, t := range needed {
				if cmd, ok := aptCmd(t); ok {
					plan.Commands = append(plan.Commands, cmd)
				} else {
					plan.Manual = append(plan.Manual, manualURL(t))
				}
			}
		} else if has("dnf") {
			for _, t := range needed {
				if cmd, ok := dnfCmd(t); ok {
					plan.Commands = append(plan.Commands, cmd)
				} else {
					plan.Manual = append(plan.Manual, manualURL(t))
				}
			}
		} else if has("pacman") {
			// refresh
			plan.Commands = append(plan.Commands, installCommand{Name: "sudo", Args: []string{"pacman", "-Sy"}})
			for _, t := range needed {
				if cmd, ok := pacmanCmd(t); ok {
					plan.Commands = append(plan.Commands, cmd)
				} else {
					plan.Manual = append(plan.Manual, manualURL(t))
				}
			}
		} else {
			for _, t := range needed {
				plan.Manual = append(plan.Manual, manualURL(t))
			}
		}
	}

	return plan
}

// --- per-manager mappings ---

func wingetCmd(tool string) (installCommand, bool) {
	switch tool {
	case "ffmpeg":
		return installCommand{Name: "winget", Args: []string{"install", "-e", "--id", "Gyan.FFmpeg"}}, true
	case "flyctl":
		return installCommand{Name: "winget", Args: []string{"install", "-e", "--id", "Fly.io.Flyctl"}}, true
	case "jpegoptim":
		// Use scoop/choco if winget lacks it
		return installCommand{}, false
	case "oxipng":
		return installCommand{}, false
	case "gifsicle":
		return installCommand{}, false
	}
	return installCommand{}, false
}

func chocoCmd(tool string) (installCommand, bool) {
	switch tool {
	case "ffmpeg":
		return installCommand{Name: "choco", Args: []string{"install", "-y", "ffmpeg"}}, true
	case "flyctl":
		// community packages may be outdated; prefer winget. If present:
		return installCommand{}, false
	case "jpegoptim":
		return installCommand{Name: "choco", Args: []string{"install", "-y", "jpegoptim"}}, true
	case "oxipng":
		return installCommand{Name: "choco", Args: []string{"install", "-y", "oxipng"}}, true
	case "gifsicle":
		return installCommand{Name: "choco", Args: []string{"install", "-y", "gifsicle"}}, true
	}
	return installCommand{}, false
}

func scoopCmd(tool string) (installCommand, bool) {
	// Ensure scoop buckets for extras if needed
	switch tool {
	case "ffmpeg":
		return installCommand{Name: "scoop", Args: []string{"install", "ffmpeg"}}, true
	case "flyctl":
		return installCommand{Name: "scoop", Args: []string{"install", "flyctl"}}, true
	case "jpegoptim":
		return installCommand{Name: "scoop", Args: []string{"install", "jpegoptim"}}, true
	case "oxipng":
		return installCommand{Name: "scoop", Args: []string{"install", "oxipng"}}, true
	case "gifsicle":
		return installCommand{Name: "scoop", Args: []string{"install", "gifsicle"}}, true
	}
	return installCommand{}, false
}

func brewCmd(tool string) (installCommand, bool) {
	switch tool {
	case "ffmpeg":
		return installCommand{Name: "brew", Args: []string{"install", "ffmpeg"}}, true
	case "flyctl":
		return installCommand{Name: "brew", Args: []string{"install", "flyctl"}}, true
	case "jpegoptim":
		return installCommand{Name: "brew", Args: []string{"install", "jpegoptim"}}, true
	case "oxipng":
		return installCommand{Name: "brew", Args: []string{"install", "oxipng"}}, true
	case "gifsicle":
		return installCommand{Name: "brew", Args: []string{"install", "gifsicle"}}, true
	}
	return installCommand{}, false
}

func aptCmd(tool string) (installCommand, bool) {
	switch tool {
	case "ffmpeg":
		return installCommand{Name: "sudo", Args: []string{"apt-get", "install", "-y", "ffmpeg"}}, true
	case "flyctl":
		// Official script is recommended
		return installCommand{Name: "bash", Args: []string{"-lc", "curl -L https://fly.io/install.sh | sh"}}, true
	case "jpegoptim":
		return installCommand{Name: "sudo", Args: []string{"apt-get", "install", "-y", "jpegoptim"}}, true
	case "oxipng":
		return installCommand{Name: "sudo", Args: []string{"apt-get", "install", "-y", "oxipng"}}, true
	case "gifsicle":
		return installCommand{Name: "sudo", Args: []string{"apt-get", "install", "-y", "gifsicle"}}, true
	}
	return installCommand{}, false
}

func dnfCmd(tool string) (installCommand, bool) {
	switch tool {
	case "ffmpeg":
		return installCommand{Name: "sudo", Args: []string{"dnf", "install", "-y", "ffmpeg"}}, true
	case "flyctl":
		return installCommand{Name: "bash", Args: []string{"-lc", "curl -L https://fly.io/install.sh | sh"}}, true
	case "jpegoptim":
		return installCommand{Name: "sudo", Args: []string{"dnf", "install", "-y", "jpegoptim"}}, true
	case "oxipng":
		return installCommand{Name: "sudo", Args: []string{"dnf", "install", "-y", "oxipng"}}, true
	case "gifsicle":
		return installCommand{Name: "sudo", Args: []string{"dnf", "install", "-y", "gifsicle"}}, true
	}
	return installCommand{}, false
}

func pacmanCmd(tool string) (installCommand, bool) {
	switch tool {
	case "ffmpeg":
		return installCommand{Name: "sudo", Args: []string{"pacman", "-S", "--noconfirm", "ffmpeg"}}, true
	case "flyctl":
		return installCommand{Name: "bash", Args: []string{"-lc", "curl -L https://fly.io/install.sh | sh"}}, true
	case "jpegoptim":
		return installCommand{Name: "sudo", Args: []string{"pacman", "-S", "--noconfirm", "jpegoptim"}}, true
	case "oxipng":
		return installCommand{Name: "sudo", Args: []string{"pacman", "-S", "--noconfirm", "oxipng"}}, true
	case "gifsicle":
		return installCommand{Name: "sudo", Args: []string{"pacman", "-S", "--noconfirm", "gifsicle"}}, true
	}
	return installCommand{}, false
}

func manualURL(tool string) string {
	switch tool {
	case "ffmpeg":
		return "FFmpeg: https://ffmpeg.org/download.html"
	case "flyctl":
		return "flyctl: https://fly.io/docs/flyctl/install/"
	case "jpegoptim":
		return "jpegoptim: https://github.com/tjko/jpegoptim"
	case "oxipng":
		return "oxipng: https://github.com/shssoichiro/oxipng"
	case "gifsicle":
		return "gifsicle: https://www.lcdf.org/gifsicle/"
	}
	return tool
}
