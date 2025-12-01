package cmd

import (
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

var installCmd = &cobra.Command{
    Use:   "install",
    Short: "Bootstrap the project: deps, tools, styles, env, and optional git init",
    Long:  "Install project dependencies, ensure required tools, scaffold styles and static assets, create .env files, and optionally initialize git.",
    RunE: func(cmd *cobra.Command, args []string) error {
        banner()
        fmt.Println("Install")
        fmt.Println("  • Ensuring Go modules...")
        if err := execx.Run(context.Background(), "go mod tidy", "go", "mod", "tidy"); err != nil {
            return fmt.Errorf("go mod tidy failed: %w", err)
        }

        if os.Getenv("GFORGE_SKIP_TOOLS") == "" && !installSkipTools {
            fmt.Println("  • Installing tools: templ, air, gotailwindcss")
            if _, err := ensureTool("templ", "github.com/a-h/templ/cmd/templ@latest"); err != nil { return err }
            if _, err := ensureTool("air", "github.com/air-verse/air@latest"); err != nil { return err }
            if _, err := ensureTool("gotailwindcss", "github.com/gotailwindcss/tailwind/cmd/gotailwindcss@latest"); err != nil { return err }
        } else {
            fmt.Println("  • Skipping tool installation (GFORGE_SKIP_TOOLS or --skip-tools)")
        }

        fmt.Println("  • Scaffolding styles")
        inputCSS := filepath.Join("app", "styles", "tailwind.input.css")
        css := "@import \"tailwindcss\" source(none);\n@source \"./app/**/*.{templ,go,html}\";\n@plugin \"./daisyui.js\";\n"
        if err := execx.WriteFileIfMissing(inputCSS, []byte(css), 0o644); err != nil {
            return err
        }

        fmt.Println("  • Fetching daisyUI plugin (optional)")
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        daisy := filepath.Join("app", "styles", "daisyui.js")
        if _, err := os.Stat(daisy); os.IsNotExist(err) {
            _ = execx.Download(ctx, "https://github.com/saadeghi/daisyui/releases/latest/download/daisyui.js", daisy)
        }
        daisyTheme := filepath.Join("app", "styles", "daisyui-theme.js")
        if _, err := os.Stat(daisyTheme); os.IsNotExist(err) {
            _ = execx.Download(ctx, "https://github.com/saadeghi/daisyui/releases/latest/download/daisyui-theme.js", daisyTheme)
        }

        fmt.Println("  • Creating static assets")
        fav := filepath.Join("app", "static", "favicon.svg")
        favicon := "<svg xmlns=\"http://www.w3.org/2000/svg\" viewBox=\"0 0 24 24\"><circle cx=\"12\" cy=\"12\" r=\"10\" fill=\"#4b5563\"/><text x=\"12\" y=\"16\" text-anchor=\"middle\" font-size=\"10\" fill=\"white\">GF</text></svg>\n"
        if err := execx.WriteFileIfMissing(fav, []byte(favicon), 0o644); err != nil {
            return err
        }

        fmt.Println("  • Creating .env.example and .env if missing")
        envExamplePath := filepath.Join(".env.example")
        if _, err := os.Stat(envExamplePath); os.IsNotExist(err) {
            if err := writeEnvExample(envExamplePath); err != nil { return err }
        }
        
        // Create .env by copying .env.example if it doesn't exist
        envPath := filepath.Join(".env")
        if _, err := os.Stat(envPath); os.IsNotExist(err) {
            // Copy from .env.example
            if exampleContent, err := os.ReadFile(envExamplePath); err == nil {
                if err := os.WriteFile(envPath, exampleContent, 0o644); err != nil { return err }
            } else {
                // Fallback: create minimal .env
                const minimal = "APP_ENV=development\nSITE_BASE_URL=http://127.0.0.1:8080\nJWT_SECRET=\n"
                if err := os.WriteFile(envPath, []byte(minimal), 0o644); err != nil { return err }
            }
        }
        
        // Ensure JWT_SECRET has a value in .env (not just empty)
        if b, err := os.ReadFile(envPath); err == nil {
            content := string(b)
            // Check if JWT_SECRET is missing or empty
            needsSecret := !strings.Contains(content, "JWT_SECRET=") || 
                          strings.Contains(content, "JWT_SECRET=\n") ||
                          strings.Contains(content, "JWT_SECRET=\r\n")
            
            if needsSecret {
                sec := genHex(32)
                lines := strings.Split(content, "\n")
                updated := false
                for i, line := range lines {
                    if strings.HasPrefix(strings.TrimSpace(line), "JWT_SECRET=") {
                        lines[i] = "JWT_SECRET=" + sec
                        updated = true
                        break
                    }
                }
                if updated {
                    _ = os.WriteFile(envPath, []byte(strings.Join(lines, "\n")), 0o644)
                } else {
                    // Append if not found
                    f, err := os.OpenFile(envPath, os.O_APPEND|os.O_WRONLY, 0o644)
                    if err == nil {
                        defer f.Close()
                        _, _ = f.WriteString("\nJWT_SECRET=" + sec + "\n")
                    }
                }
            }
        }

        fmt.Println("  • Creating sitemap registry (app/sitemap/urls.txt) if missing")
        sitemapDir := filepath.Join("app", "sitemap")
        sitemapFile := filepath.Join(sitemapDir, "urls.txt")
        if _, err := os.Stat(sitemapFile); os.IsNotExist(err) {
            if err := os.MkdirAll(sitemapDir, 0o755); err != nil { return err }
            content := "# Add one path or absolute URL per line.\n# Lines starting with # are ignored.\n/\n"
            if err := os.WriteFile(sitemapFile, []byte(content), 0o644); err != nil { return err }
        }

        // Create Dockerfile and .dockerignore if missing (for containerized deployments)
        if err := ensureDockerFiles(); err != nil {
            fmt.Println("  ⚠️  Warning: Could not create Docker files:", err)
        }

        if os.Getenv("GFORGE_SKIP_TOOLS") == "" && !installSkipTools {
            fmt.Println("  • Running initial build: templ generate & tailwind build")
            if templPath, err := ensureTool("templ", "github.com/a-h/templ/cmd/templ@latest"); err == nil {
                _ = execx.Run(context.Background(), "templ", templPath, "generate", "-include-version=false", "-include-timestamp=false")
            }
            if gwPath, err := ensureTool("gotailwindcss", "github.com/gotailwindcss/tailwind/cmd/gotailwindcss@latest"); err == nil {
                _ = execx.Run(context.Background(), "gotailwindcss build", gwPath, "build", "-o", "./app/styles/output.css", "./app/styles/tailwind.input.css")
            }
        }

        if installGitInit {
            if _, ok := execx.Look("git"); !ok {
                fmt.Println("  • Git not found; skipping repository initialization")
                fmt.Println("    Install git to use --git flag:")
                printGitInstallHelp()
            } else {
                fmt.Println("  • Initializing git repository")
                // best-effort: ignore errors to keep install idempotent
                _ = execx.Run(context.Background(), "git init", "git", "init")
                _ = execx.Run(context.Background(), "git add .", "git", "add", ".")
                _ = execx.Run(context.Background(), "git commit", "git", "commit", "-m", "chore(install): bootstrap project")
            }
        }

        fmt.Println("────────────────────────────────────────")
        fmt.Println("Install complete.")
        
        // Check for external deployment tools and provide guidance
        fmt.Println("")
        fmt.Println("🔍 Checking deployment tools (optional):")
        checkExternalTool("Cloudflare Wrangler", "wrangler", "npm install -g wrangler", "https://developers.cloudflare.com/workers/wrangler/install-and-update/")
        checkExternalTool("Docker", "docker", "", "https://docs.docker.com/get-docker/")
        
        fmt.Println("")
        fmt.Println("💡 Gothic Forge Opinionated Stack:")
        fmt.Println("   • Cloudflare Pages + Functions - static hosting + edge compute")
        fmt.Println("   • Leapcell - Go backend compute (requires Docker)")
        fmt.Println("   • CockroachDB Serverless - PostgreSQL database")
        fmt.Println("   • Aiven Valkey - Redis-compatible cache")
        fmt.Println("")
        fmt.Println("✅ Run 'gforge doctor' to see full system check")
        fmt.Println("✅ Run 'gforge dev' to start development server")
        
        return nil
    },
}

var (
    installSkipTools bool
    installGitInit   bool
)

func init() {
    installCmd.Flags().BoolVar(&installSkipTools, "skip-tools", false, "skip installing CLI tools (templ, air, gotailwindcss)")
    installCmd.Flags().BoolVar(&installGitInit, "git", false, "initialize a git repository and make initial commit")
    rootCmd.AddCommand(installCmd)
}

// genHex returns a random hex string with n bytes.
func genHex(n int) string {
    b := make([]byte, n)
    if _, err := rand.Read(b); err != nil { return "" }
    return hex.EncodeToString(b)
}

// checkExternalTool checks if a tool is installed and provides installation guidance if missing.
func checkExternalTool(name, command, installCmd, docsURL string) {
    _, err := exec.LookPath(command)
    if err != nil {
        fmt.Printf("   ⚠️  %s: not found\n", name)
        if installCmd != "" {
            fmt.Printf("      Install: %s\n", installCmd)
        }
        if docsURL != "" {
            fmt.Printf("      Docs: %s\n", docsURL)
        }
    } else {
        fmt.Printf("   ✅ %s: installed\n", name)
    }
}

// ensureDockerFiles creates Dockerfile and .dockerignore if they don't exist.
// These files are essential for containerized deployments (Leapcell, etc.).
func ensureDockerFiles() error {
    dockerfilePath := "Dockerfile"
    dockerignorePath := ".dockerignore"
    
    // Check if Dockerfile exists
    if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
        fmt.Println("  • Creating Dockerfile for container deployments")
        if err := createDockerfile(dockerfilePath); err != nil {
            return fmt.Errorf("failed to create Dockerfile: %w", err)
        }
        fmt.Println("    → Dockerfile created with multi-stage build")
    } else {
        fmt.Println("  • Dockerfile already exists")
    }
    
    // Check if .dockerignore exists
    if _, err := os.Stat(dockerignorePath); os.IsNotExist(err) {
        fmt.Println("  • Creating .dockerignore for optimized builds")
        if err := createDockerignore(dockerignorePath); err != nil {
            return fmt.Errorf("failed to create .dockerignore: %w", err)
        }
        fmt.Println("    → .dockerignore created")
    } else {
        fmt.Println("  • .dockerignore already exists")
    }
    
    return nil
}

// createDockerfile writes the production-ready Dockerfile template.
func createDockerfile(path string) error {
    content := `# Gothic Forge v3 - Production Dockerfile
#
# This Dockerfile uses multi-stage builds to:
# 1. Keep the final image small (security + faster deployments)
# 2. Separate build-time dependencies from runtime
# 3. Follow Docker best practices and least-privilege principles
#
# Educational Resources:
# - Multi-stage builds: https://docs.docker.com/build/building/multi-stage/
# - Security best practices: https://docs.docker.com/develop/security-best-practices/

# ═══════════════════════════════════════════════════════════════════════════
# Stage 1: Builder
# ═══════════════════════════════════════════════════════════════════════════
# Why golang:alpine? 
# - Alpine Linux is minimal (~5MB base) vs debian (~124MB)
# - Includes go toolchain for building
# - Smaller attack surface (fewer packages = fewer vulnerabilities)
FROM golang:1.24-alpine AS builder

# Install build dependencies
# - git: Required for go mod download with private repos
# - ca-certificates: Required for HTTPS requests during build
# - tzdata: Timezone data (copied to final image)
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go.mod and go.sum first (Docker layer caching optimization)
# Why? If dependencies haven't changed, Docker reuses this layer
# This makes rebuilds much faster when only code changes
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

# Copy the entire source code
COPY . .

# Build the application
# Flags explained:
# -ldflags="-s -w": Strip debug info and symbol table (reduces binary size by ~30%)
# -trimpath: Remove file system paths from binary (security + reproducible builds)
# CGO_ENABLED=0: Disable CGO for fully static binary (portable across Linux distros)
# GOOS=linux: Target Linux (even if building on Windows/macOS)
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -trimpath \
    -o server \
    ./cmd/server

# ═══════════════════════════════════════════════════════════════════════════
# Stage 2: Runtime
# ═══════════════════════════════════════════════════════════════════════════
# Why scratch/alpine?
# Option A (scratch): Absolute minimal (~binary only, 15-20MB total)
# Option B (alpine): Minimal Linux with shell (~30MB total, easier debugging)
# We choose alpine for production debugging capabilities
FROM alpine:latest

# Install runtime dependencies (minimal)
# - ca-certificates: HTTPS certificate validation
# - tzdata: Timezone support (for logs, timestamps)
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user for security
# Why? Running as root is a security risk. If the container is compromised,
# attacker has root privileges. Non-root limits damage.
# UID 1000 is standard for non-root users
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Create necessary directories with proper permissions
RUN mkdir -p /app/app/static /app/app/styles /app/app/db && \
    chown -R appuser:appuser /app

# Set working directory
WORKDIR /app

# Copy compiled binary from builder stage
COPY --from=builder /build/server /app/server

# Copy application assets
# These are needed for Templ templates, static files, and database migrations
COPY --chown=appuser:appuser app/static ./app/static
COPY --chown=appuser:appuser app/styles ./app/styles
COPY --chown=appuser:appuser app/db ./app/db

# Copy timezone data for consistent timestamps across deployments
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Switch to non-root user (all subsequent commands run as this user)
USER appuser

# Expose port (documentation only, doesn't actually publish the port)
# The application should read HTTP_PORT from environment
EXPOSE 8080

# Health check (Docker/Kubernetes use this to determine container health)
# Why /readyz? It checks if the app is ready to serve traffic (DB connected, etc.)
# Interval: Check every 30s
# Timeout: Wait max 3s for response
# Start period: Wait 10s after container starts before first check
# Retries: Mark unhealthy after 3 consecutive failures
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/readyz || exit 1

# Environment variables (defaults, override in deployment)
ENV APP_ENV=production \
    HTTP_HOST=0.0.0.0 \
    HTTP_PORT=8080 \
    LOG_FORMAT=json

# Run the application
# Why array form ["cmd", "arg"]? 
# - Proper signal handling (SIGTERM for graceful shutdown)
# - No shell intermediary (more secure, faster startup)
CMD ["/app/server"]
`
    return os.WriteFile(path, []byte(content), 0o644)
}

// createDockerignore writes the .dockerignore template.
func createDockerignore(path string) error {
    content := `# Gothic Forge v3 - Docker Build Context Exclusions
#
# Why .dockerignore?
# 1. Faster builds: Smaller build context = faster upload to Docker daemon
# 2. Security: Don't accidentally include secrets or sensitive files
# 3. Smaller images: Prevents unnecessary files in final image layers
#
# Similar to .gitignore but for Docker builds

# ═══════════════════════════════════════════════════════════════════════════
# Version Control
# ═══════════════════════════════════════════════════════════════════════════
.git/
.github/
.gitignore
.gitattributes

# ═══════════════════════════════════════════════════════════════════════════
# Documentation (not needed in runtime image)
# ═══════════════════════════════════════════════════════════════════════════
README.md
CONTRIBUTING.md
CODE_OF_CONDUCT.md
SECURITY.md
LICENSE
PHILOSOPHY.md
*.md

# ═══════════════════════════════════════════════════════════════════════════
# Environment & Secrets (CRITICAL: Never include in Docker image)
# ═══════════════════════════════════════════════════════════════════════════
.env
.env.*
!.env.example
*.key
*.pem
*.crt
*.cert
secrets/

# ═══════════════════════════════════════════════════════════════════════════
# Development Tools & Config
# ═══════════════════════════════════════════════════════════════════════════
.vscode/
.idea/
.editorconfig
.air.toml
.golangci.yml
.goreleaser.yaml

# ═══════════════════════════════════════════════════════════════════════════
# Build Artifacts & Binaries (rebuilt in Docker)
# ═══════════════════════════════════════════════════════════════════════════
dist/
build/
*.exe
*.dll
*.so
*.dylib
gforge
gforge.exe
server
server.exe

# ═══════════════════════════════════════════════════════════════════════════
# Test Files & Coverage (not needed in production)
# ═══════════════════════════════════════════════════════════════════════════
tests/
*_test.go
*.test
coverage.out
coverage.html

# ═══════════════════════════════════════════════════════════════════════════
# Temporary & Cache Files
# ═══════════════════════════════════════════════════════════════════════════
tmp/
temp/
*.tmp
*.cache
*.log

# ═══════════════════════════════════════════════════════════════════════════
# OS-specific Files
# ═══════════════════════════════════════════════════════════════════════════
.DS_Store
Thumbs.db
desktop.ini

# ═══════════════════════════════════════════════════════════════════════════
# Node modules (if accidentally present)
# ═══════════════════════════════════════════════════════════════════════════
node_modules/
npm-debug.log
yarn-error.log

# ═══════════════════════════════════════════════════════════════════════════
# Docker files (don't need Dockerfile inside container)
# ═══════════════════════════════════════════════════════════════════════════
Dockerfile*
.dockerignore
docker-compose*.yml

# ═══════════════════════════════════════════════════════════════════════════
# CI/CD configs (not needed at runtime)
# ═══════════════════════════════════════════════════════════════════════════
.github/
Jenkinsfile
.travis.yml
.gitlab-ci.yml

# ═══════════════════════════════════════════════════════════════════════════
# Deployment configs (managed outside container)
# ═══════════════════════════════════════════════════════════════════════════
fly.toml
render.yaml
Caddyfile

# ═══════════════════════════════════════════════════════════════════════════
# IMPORTANT: What we DO include
# ═══════════════════════════════════════════════════════════════════════════
# ✓ go.mod, go.sum (dependencies)
# ✓ cmd/ (source code)
# ✓ internal/ (source code)
# ✓ app/templates/ (Templ templates)
# ✓ app/static/ (static assets)
# ✓ app/styles/ (CSS)
# ✓ app/db/ (migrations)
`
    return os.WriteFile(path, []byte(content), 0o644)
}
