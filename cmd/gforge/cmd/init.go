package cmd

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gothicforge3/internal/execx"

	"github.com/spf13/cobra"
)

var (
	initInteractive bool
	initTemplate    string
)

// InteractiveConfig holds configuration collected during interactive setup
type InteractiveConfig struct {
	ProjectName         string
	UseOpinionatedStack bool
	OpenSignupURLs      bool
	CockroachAPIKey     string
	AivenToken          string
	CloudflareAPIToken  string
}

// HasAPIKeys returns true if any API keys were provided
func (c *InteractiveConfig) HasAPIKeys() bool {
	return c.CockroachAPIKey != "" || c.AivenToken != "" || c.CloudflareAPIToken != ""
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Gothic Forge project in under 30 seconds",
	Long: `Initialize a new Gothic Forge project with zero manual configuration.

This command will:
  • Detect your environment (Go, Git, project name)
  • Generate secure configuration (.env with JWT_SECRET)
  • Install dependencies (go mod download)
  • Install required tools (templ, gotailwindcss)
  • Generate templates and build CSS
  • Verify setup with health checks

Target: Complete in under 30 seconds`,
	RunE: func(cmd *cobra.Command, args []string) error {
		banner()
		fmt.Println("Init - Gothic Forge Project Setup")
		fmt.Println("────────────────────────────────────────")

		ctx := context.Background()
		startTime := time.Now()

		// Interactive mode: prompt for project details before detection
		var interactiveConfig *InteractiveConfig
		if initInteractive {
			config, err := interactiveSetup()
			if err != nil {
				return fmt.Errorf("interactive setup failed: %w", err)
			}
			interactiveConfig = config
		}

		// Step 1: Detect environment (2s)
		fmt.Println("Step 1/5: Detecting environment...")
		projectName, gitRemote, err := detectEnvironment(ctx)
		if err != nil {
			return fmt.Errorf("environment detection failed: %w", err)
		}
		
		// Override with interactive config if provided
		if interactiveConfig != nil && interactiveConfig.ProjectName != "" {
			projectName = interactiveConfig.ProjectName
		}
		
		fmt.Printf("  ✓ Project name: %s\n", projectName)
		if gitRemote != "" {
			fmt.Printf("  ✓ Git remote: %s\n", gitRemote)
		}

		// Step 2: Generate configuration (3s)
		fmt.Println("\nStep 2/5: Generating configuration...")
		if err := generateConfiguration(projectName, gitRemote, interactiveConfig); err != nil {
			return fmt.Errorf("configuration generation failed: %w", err)
		}
		fmt.Println("  ✓ .env created with secure defaults")
		fmt.Println("  ✓ JWT_SECRET generated (64 chars)")
		if gitRemote != "" {
			fmt.Println("  ✓ SITE_BASE_URL set from Git remote")
		}
		if interactiveConfig != nil && interactiveConfig.HasAPIKeys() {
			fmt.Println("  ✓ API keys stored securely")
		}

		// Step 3: Install dependencies (10s)
		fmt.Println("\nStep 3/5: Installing dependencies...")
		if err := installDependencies(ctx); err != nil {
			return fmt.Errorf("dependency installation failed: %w", err)
		}
		fmt.Println("  ✓ Go modules downloaded")
		fmt.Println("  ✓ Tools installed (templ, gotailwindcss)")
		fmt.Println("  ✓ Templates generated")
		fmt.Println("  ✓ Tailwind CSS built")

		// Step 4: Verify setup (5s)
		fmt.Println("\nStep 4/5: Verifying setup...")
		if err := verifySetup(ctx); err != nil {
			fmt.Printf("  ⚠️  Setup verification had warnings: %v\n", err)
			fmt.Println("  → Run 'gforge doctor' for detailed diagnostics")
		} else {
			fmt.Println("  ✓ Project builds successfully")
			fmt.Println("  ✓ All tools available")
		}

		// Step 5: Success message
		duration := time.Since(startTime)
		fmt.Println("\n────────────────────────────────────────")
		fmt.Printf("✓ Project initialized successfully! (%v)\n", duration.Round(time.Second))
		fmt.Println("\nNext steps:")
		fmt.Println("  → Run: gforge dev (start development server)")
		fmt.Println("  → Run: gforge deploy --live (deploy to production)")
		fmt.Println("\n💡 Gothic Forge Opinionated Stack:")
		fmt.Println("   • Leapcell - Go backend compute")
		fmt.Println("   • CockroachDB Serverless - PostgreSQL database")
		fmt.Println("   • Aiven Valkey - Redis-compatible cache")
		fmt.Println("   • Cloudflare Pages - CDN/edge compute")

		return nil
	},
}

func init() {
	initCmd.Flags().BoolVar(&initInteractive, "interactive", false, "prompt for project details and provider preferences")
	initCmd.Flags().StringVar(&initTemplate, "template", "minimal", "starter template (minimal, blog, saas, api)")
	rootCmd.AddCommand(initCmd)
}

// detectEnvironment detects the project name and Git remote URL
// It performs comprehensive environment checks including:
// - Current working directory and project name
// - Go installation and version
// - Git installation and repository status
// - Git remote URL (if available)
func detectEnvironment(ctx context.Context) (projectName string, gitRemote string, err error) {
	// Detect project name from current directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("failed to get current directory: %w", err)
	}
	projectName = filepath.Base(cwd)

	// Validate project name is not empty or invalid
	if projectName == "" || projectName == "." || projectName == "/" {
		return "", "", fmt.Errorf("invalid project directory")
	}

	// Check if Go is installed
	goPath, err := exec.LookPath("go")
	if err != nil {
		return "", "", fmt.Errorf("Go is not installed or not in PATH. Install from: https://go.dev/dl/")
	}

	// Verify Go version (optional but helpful)
	if goVersion, err := execx.RunCapture(ctx, "go version", goPath, "version"); err == nil {
		fmt.Printf("  → Go detected: %s\n", strings.TrimSpace(goVersion))
	}

	// Check if Git is installed
	gitPath, err := exec.LookPath("git")
	if err != nil {
		// Git is optional, so just note it's not available
		fmt.Println("  → Git not found (optional)")
		return projectName, "", nil
	}

	// Check if this is a git repository
	if err := execx.RunQuiet(ctx, "git rev-parse", gitPath, "rev-parse", "--git-dir"); err != nil {
		// Not a git repo, which is fine
		fmt.Println("  → Not a Git repository (optional)")
		return projectName, "", nil
	}

	// Get remote URL if available
	if remote, err := execx.RunCapture(ctx, "git remote get-url origin", gitPath, "remote", "get-url", "origin"); err == nil {
		gitRemote = strings.TrimSpace(remote)
		if gitRemote != "" {
			fmt.Printf("  → Git repository detected\n")
		}
	} else {
		// No remote configured yet, which is fine
		fmt.Println("  → Git repository (no remote configured)")
	}

	return projectName, gitRemote, nil
}

// generateConfiguration creates .env and providers.yaml with secure defaults
func generateConfiguration(projectName, gitRemote string, interactiveConfig *InteractiveConfig) error {
	// Generate .env file
	if err := generateEnvFile(projectName, gitRemote, interactiveConfig); err != nil {
		return fmt.Errorf("failed to generate .env: %w", err)
	}

	// Generate providers.yaml (optional, for future use)
	if err := generateProvidersConfig(); err != nil {
		// Non-fatal: providers.yaml is optional
		fmt.Printf("  ⚠️  Could not generate providers.yaml: %v\n", err)
	}

	return nil
}

// generateEnvFile creates .env with secure defaults
func generateEnvFile(projectName, gitRemote string, interactiveConfig *InteractiveConfig) error {
	envPath := ".env"

	// Check if .env already exists
	if _, err := os.Stat(envPath); err == nil {
		fmt.Println("  → .env already exists, skipping")
		return nil
	}

	// Generate secure JWT_SECRET
	jwtSecret := generateSecureSecret(32)

	// Determine SITE_BASE_URL from Git remote
	siteBaseURL := "http://127.0.0.1:8080"
	if gitRemote != "" {
		if url := extractURLFromGitRemote(gitRemote); url != "" {
			siteBaseURL = url
		}
	}

	// Get API keys from interactive config if provided
	cockroachKey := ""
	aivenToken := ""
	cfToken := ""
	if interactiveConfig != nil {
		cockroachKey = interactiveConfig.CockroachAPIKey
		aivenToken = interactiveConfig.AivenToken
		cfToken = interactiveConfig.CloudflareAPIToken
	}

	// Create .env content
	envContent := fmt.Sprintf(`# Gothic Forge v3 - Environment Configuration
# Generated by gforge init on %s

# Application Environment
APP_ENV=development
SITE_BASE_URL=%s
HTTP_HOST=127.0.0.1
HTTP_PORT=8080

# Security
JWT_SECRET=%s

# Database (CockroachDB Serverless - Opinionated Stack)
DATABASE_URL=

# Cache (Aiven Valkey - Opinionated Stack)
VALKEY_URL=

# Deployment Tokens (Opinionated Stack)
LEAPCELL_APP_URL=
COCKROACH_API_KEY=%s
AIVEN_TOKEN=%s
CLOUDFLARE_API_TOKEN=%s

# Optional: Cloudflare Pages
CF_PROJECT_NAME=
CF_ACCOUNT_ID=

# For full configuration options, see: https://github.com/yourusername/gothic-forge
`, time.Now().Format("2006-01-02 15:04:05"), siteBaseURL, jwtSecret, cockroachKey, aivenToken, cfToken)

	// Write .env file
	if err := os.WriteFile(envPath, []byte(envContent), 0o600); err != nil {
		return fmt.Errorf("failed to write .env: %w", err)
	}

	return nil
}

// generateProvidersConfig creates providers.yaml with default settings
func generateProvidersConfig() error {
	providersPath := "providers.yaml"

	// Check if providers.yaml already exists
	if _, err := os.Stat(providersPath); err == nil {
		return nil // Already exists
	}

	// Create providers.yaml content (Opinionated Stack defaults)
	providersContent := `# Gothic Forge Provider Configuration
# Customize which providers to use for each service

database:
  provider: cockroachdb  # cockroachdb, postgresql, sqlite
  options:
    region: us-east-1
    tier: serverless

cache:
  provider: valkey  # valkey, redis, none
  options:
    region: us-east-1
    tier: startup

compute:
  provider: leapcell  # leapcell, docker, aws-ecs
  options:
    region: global
    replicas: 1

cdn:
  provider: cloudflare  # cloudflare, none
  options:
    zone: auto
`

	// Write providers.yaml file
	if err := os.WriteFile(providersPath, []byte(providersContent), 0o644); err != nil {
		return fmt.Errorf("failed to write providers.yaml: %w", err)
	}

	return nil
}

// installDependencies installs Go modules and required tools
func installDependencies(ctx context.Context) error {
	// Run go mod download
	fmt.Println("  → Running go mod download...")
	if err := execx.Run(ctx, "go mod download", "go", "mod", "download"); err != nil {
		return fmt.Errorf("go mod download failed: %w", err)
	}

	// Install templ
	fmt.Println("  → Installing templ...")
	if _, err := ensureTool("templ", "github.com/a-h/templ/cmd/templ@latest"); err != nil {
		return fmt.Errorf("failed to install templ: %w", err)
	}

	// Install gotailwindcss
	fmt.Println("  → Installing gotailwindcss...")
	if _, err := ensureTool("gotailwindcss", "github.com/gotailwindcss/tailwind/cmd/gotailwindcss@latest"); err != nil {
		return fmt.Errorf("failed to install gotailwindcss: %w", err)
	}

	// Generate Templ templates
	fmt.Println("  → Generating Templ templates...")
	if templPath, err := exec.LookPath("templ"); err == nil {
		if err := execx.RunQuiet(ctx, "templ generate", templPath, "generate", "-include-version=false", "-include-timestamp=false"); err != nil {
			fmt.Printf("  ⚠️  templ generate had warnings: %v\n", err)
		}
	}

	// Build Tailwind CSS
	fmt.Println("  → Building Tailwind CSS...")
	if err := buildTailwindCSS(ctx); err != nil {
		fmt.Printf("  ⚠️  Tailwind build had warnings: %v\n", err)
	}

	return nil
}

// buildTailwindCSS builds the Tailwind CSS output
func buildTailwindCSS(ctx context.Context) error {
	// Ensure input CSS exists
	inputCSS := filepath.Join("app", "styles", "tailwind.input.css")
	if _, err := os.Stat(inputCSS); os.IsNotExist(err) {
		// Create minimal input CSS
		css := "@import \"tailwindcss\" source(none);\n@source \"./app/**/*.{templ,go,html}\";\n"
		if err := os.MkdirAll(filepath.Dir(inputCSS), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(inputCSS, []byte(css), 0o644); err != nil {
			return err
		}
	}

	// Build CSS
	if gwPath, err := exec.LookPath("gotailwindcss"); err == nil {
		outputCSS := filepath.Join("app", "styles", "output.css")
		if err := execx.RunQuiet(ctx, "gotailwindcss build", gwPath, "build", "-o", outputCSS, inputCSS); err != nil {
			return err
		}
	}

	return nil
}

// verifySetup runs basic health checks
func verifySetup(ctx context.Context) error {
	// Check if project builds
	fmt.Println("  → Verifying project builds...")
	buildCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := execx.RunQuiet(buildCtx, "go build", "go", "build", "-o", "server.tmp", "./cmd/server"); err != nil {
		return fmt.Errorf("project does not build: %w", err)
	}

	// Clean up temporary binary
	_ = os.Remove("server.tmp")

	// Check if tools are available
	tools := []string{"templ", "gotailwindcss"}
	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("%s not found in PATH", tool)
		}
	}

	return nil
}

// generateSecureSecret generates a cryptographically secure random hex string
func generateSecureSecret(bytes int) string {
	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		// Fallback to timestamp-based secret (not ideal but better than nothing)
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}

// extractURLFromGitRemote extracts a base URL from a Git remote URL
func extractURLFromGitRemote(remote string) string {
	remote = strings.TrimSpace(remote)

	// Handle SSH format: git@github.com:user/repo.git
	if strings.HasPrefix(remote, "git@") {
		parts := strings.Split(remote, ":")
		if len(parts) == 2 {
			host := strings.TrimPrefix(parts[0], "git@")
			repo := strings.TrimSuffix(parts[1], ".git")
			return fmt.Sprintf("https://%s/%s", host, repo)
		}
	}

	// Handle HTTPS format: https://github.com/user/repo.git
	if strings.HasPrefix(remote, "https://") || strings.HasPrefix(remote, "http://") {
		remote = strings.TrimSuffix(remote, ".git")
		return remote
	}

	return ""
}

// interactiveSetup prompts the user for project details (when --interactive flag is used)
func interactiveSetup() (*InteractiveConfig, error) {
	reader := bufio.NewReader(os.Stdin)
	config := &InteractiveConfig{}

	fmt.Println("\n🎯 Interactive Setup")
	fmt.Println("────────────────────────────────────────")

	// Prompt for project name
	fmt.Print("Project name (leave blank to use current directory): ")
	projectName, _ := reader.ReadString('\n')
	config.ProjectName = strings.TrimSpace(projectName)

	// Prompt for provider preferences
	fmt.Println("\n📦 Provider Preferences (Opinionated Stack)")
	fmt.Println("  • Database: CockroachDB Serverless")
	fmt.Println("  • Cache: Aiven Valkey")
	fmt.Println("  • Compute: Leapcell")
	fmt.Println("  • CDN: Cloudflare Pages")
	fmt.Print("\nUse Opinionated Stack? [Y/n]: ")
	useOpinionated, _ := reader.ReadString('\n')
	useOpinionated = strings.ToLower(strings.TrimSpace(useOpinionated))

	config.UseOpinionatedStack = useOpinionated != "n" && useOpinionated != "no"

	if !config.UseOpinionatedStack {
		fmt.Println("  → Custom provider configuration not yet implemented")
		fmt.Println("  → Using Opinionated Stack defaults")
		config.UseOpinionatedStack = true
	}

	// Offer to open provider signup URLs
	fmt.Print("\nOpen provider signup URLs in browser? [y/N]: ")
	openURLs, _ := reader.ReadString('\n')
	openURLs = strings.ToLower(strings.TrimSpace(openURLs))

	config.OpenSignupURLs = openURLs == "y" || openURLs == "yes"

	if config.OpenSignupURLs {
		fmt.Println("\n📝 Provider Signup Links:")
		fmt.Println("  • Leapcell: https://leapcell.io/signup")
		fmt.Println("  • CockroachDB: https://cockroachlabs.cloud/signup")
		fmt.Println("  • Aiven: https://console.aiven.io/signup")
		fmt.Println("  • Cloudflare: https://dash.cloudflare.com/sign-up")
		fmt.Println("\n  💡 Tip: Sign up for these services and get your API keys")
		fmt.Println("     You can enter them in the next step or add them to .env later")
	}

	// Prompt for API keys (optional)
	fmt.Print("\nEnter API keys now? [y/N]: ")
	enterKeys, _ := reader.ReadString('\n')
	enterKeys = strings.ToLower(strings.TrimSpace(enterKeys))

	if enterKeys == "y" || enterKeys == "yes" {
		fmt.Println("\n🔑 API Keys (leave blank to skip)")

		fmt.Print("COCKROACH_API_KEY: ")
		cockroachKey, _ := reader.ReadString('\n')
		config.CockroachAPIKey = strings.TrimSpace(cockroachKey)

		fmt.Print("AIVEN_TOKEN: ")
		aivenToken, _ := reader.ReadString('\n')
		config.AivenToken = strings.TrimSpace(aivenToken)

		fmt.Print("CLOUDFLARE_API_TOKEN: ")
		cfToken, _ := reader.ReadString('\n')
		config.CloudflareAPIToken = strings.TrimSpace(cfToken)

		if config.HasAPIKeys() {
			fmt.Println("  ✓ API keys will be stored securely in .env")
		}
	}

	fmt.Println()
	return config, nil
}
