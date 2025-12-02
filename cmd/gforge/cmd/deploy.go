package cmd

import (
	"context"
	"fmt"

	"gothicforge3/internal/env"
	"gothicforge3/internal/execx"

	"github.com/spf13/cobra"
)

var (
	deployDryRun bool
	deployCheck  bool
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy your Gothic Forge application",
	Long: `Deploy your Gothic Forge application to your chosen platform.

Gothic Forge v9.3 takes a minimal approach - you choose your deployment platform:
  • Docker + any cloud provider (AWS, GCP, Azure, DigitalOcean, etc.)
  • Platform-as-a-Service (Heroku, Render, Fly.io, Railway, etc.)
  • Serverless (AWS Lambda, Google Cloud Run, Azure Functions, etc.)
  • Traditional hosting (VPS, dedicated servers, etc.)

This command provides guidance and checks your configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		banner()
		_ = env.Load()

		if deployCheck {
			return runDeployPreflightCheck()
		}

		if deployDryRun {
			fmt.Println("\n🔍 Deploy Dry Run")
			fmt.Println("════════════════════════════════════════")
			return runDeployDryRun()
		}

		return runDeployGuide()
	},
}

func init() {
	deployCmd.Flags().BoolVar(&deployDryRun, "dry-run", false, "Show what would be deployed without making changes")
	deployCmd.Flags().BoolVar(&deployCheck, "check", false, "Check deployment prerequisites")
	rootCmd.AddCommand(deployCmd)
}

func runDeployPreflightCheck() error {
	fmt.Println("\n✓ Deployment Preflight Check")
	fmt.Println("════════════════════════════════════════")

	ctx := context.Background()

	// Check Go
	if err := execx.RunQuiet(ctx, "go", "version"); err != nil {
		fmt.Println("❌ Go not found")
		return fmt.Errorf("go is required for deployment")
	}
	fmt.Println("✓ Go installed")

	// Check Git
	if err := execx.RunQuiet(ctx, "git", "version"); err != nil {
		fmt.Println("❌ Git not found")
		return fmt.Errorf("git is required for deployment")
	}
	fmt.Println("✓ Git installed")

	// Check Docker (optional)
	if err := execx.RunQuiet(ctx, "docker", "version"); err != nil {
		fmt.Println("⚠️  Docker not found (optional)")
	} else {
		fmt.Println("✓ Docker installed")
	}

	// Check .env file
	if err := env.Load(); err != nil {
		fmt.Println("⚠️  .env file not found (optional)")
	} else {
		fmt.Println("✓ .env file found")
	}

	// Check build
	fmt.Println("\n📦 Testing build...")
	if err := execx.RunQuiet(ctx, "go", "build", "-o", "server_test", "./cmd/server"); err != nil {
		fmt.Println("❌ Build failed")
		return fmt.Errorf("build failed: %w", err)
	}
	fmt.Println("✓ Build successful")

	fmt.Println("\n✅ All checks passed!")
	fmt.Println("\nYou're ready to deploy. Choose your platform:")
	fmt.Println("  • Docker: Build and push container image")
	fmt.Println("  • PaaS: Push to Git and let platform build")
	fmt.Println("  • Serverless: Package and deploy function")
	fmt.Println("  • Traditional: Build binary and upload to server")

	return nil
}

func runDeployDryRun() error {
	fmt.Println("\n📋 Deployment Plan:")
	fmt.Println("  1. Build production binary")
	fmt.Println("  2. Run database migrations")
	fmt.Println("  3. Deploy to your chosen platform")
	fmt.Println("  4. Verify health checks")
	fmt.Println("\n💡 This is a dry run - no actual deployment will occur")
	return nil
}

func runDeployGuide() error {
	fmt.Println("\n🚀 Gothic Forge Deployment Guide")
	fmt.Println("════════════════════════════════════════")
	fmt.Println("\nGothic Forge v9.3 is platform-agnostic.")
	fmt.Println("Choose the deployment method that fits your needs:")

	fmt.Println("\n📦 Option 1: Docker (Recommended)")
	fmt.Println("────────────────────────────────────────")
	fmt.Println("Build and deploy as a container:")
	fmt.Println("  1. docker build -t myapp .")
	fmt.Println("  2. docker push myapp:latest")
	fmt.Println("  3. Deploy to any cloud provider")
	fmt.Println("\nWorks with: AWS ECS, Google Cloud Run, Azure Container")
	fmt.Println("Instances, DigitalOcean App Platform, Fly.io, Railway, etc.")

	fmt.Println("\n🌐 Option 2: Platform-as-a-Service")
	fmt.Println("────────────────────────────────────────")
	fmt.Println("Push to Git and let the platform build:")
	fmt.Println("  1. git push heroku main")
	fmt.Println("  2. Platform detects Go and builds automatically")
	fmt.Println("\nWorks with: Heroku, Render, Railway, Fly.io, etc.")

	fmt.Println("\n⚡ Option 3: Serverless")
	fmt.Println("────────────────────────────────────────")
	fmt.Println("Deploy as a serverless function:")
	fmt.Println("  1. Package your application")
	fmt.Println("  2. Deploy to serverless platform")
	fmt.Println("\nWorks with: AWS Lambda, Google Cloud Functions,")
	fmt.Println("Azure Functions, Vercel, Netlify Functions, etc.")

	fmt.Println("\n🖥️  Option 4: Traditional Hosting")
	fmt.Println("────────────────────────────────────────")
	fmt.Println("Build and upload binary:")
	fmt.Println("  1. go build -o server ./cmd/server")
	fmt.Println("  2. Upload to your server")
	fmt.Println("  3. Run with systemd or supervisor")
	fmt.Println("\nWorks with: VPS, dedicated servers, on-premise, etc.")

	fmt.Println("\n📚 Next Steps:")
	fmt.Println("────────────────────────────────────────")
	fmt.Println("  1. Choose your deployment platform")
	fmt.Println("  2. Set up your database (PostgreSQL)")
	fmt.Println("  3. Set up your cache (Redis/Valkey) - optional")
	fmt.Println("  4. Configure environment variables")
	fmt.Println("  5. Deploy your application")
	fmt.Println("  6. Run migrations: gforge db --migrate")
	fmt.Println("  7. Verify health: curl https://your-app.com/healthz")

	fmt.Println("\n💡 Tip: Run 'gforge deploy --check' to verify prerequisites")

	return nil
}
