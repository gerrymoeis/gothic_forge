package tests

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Test_Deploy_DryRun_Leapcell verifies that the deploy command runs in dry-run mode
// and displays the Opinionated Stack configuration (Leapcell + CockroachDB + Aiven + Cloudflare)
func Test_Deploy_DryRun_Leapcell(t *testing.T) {
	// Set up minimal environment for dry-run
	os.Setenv("LOG_FORMAT", "off")
	os.Setenv("COCKROACH_API_KEY", "test-cockroach-key")
	os.Setenv("JWT_SECRET", "this-is-a-very-long-secret-key-with-more-than-32-chars")
	defer func() {
		os.Unsetenv("LOG_FORMAT")
		os.Unsetenv("COCKROACH_API_KEY")
		os.Unsetenv("JWT_SECRET")
	}()

	cmd := exec.Command("go", "run", "./cmd/gforge", "deploy", "--dry-run")
	cmd.Dir = ".."
	cmd.Env = append(os.Environ(),
		"LOG_FORMAT=off",
		"COCKROACH_API_KEY=test-cockroach-key",
		"JWT_SECRET=this-is-a-very-long-secret-key-with-more-than-32-chars",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("deploy --dry-run failed: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)

	// Verify Opinionated Stack components are mentioned
	requiredStrings := []string{
		"Opinionated Stack",
		"Leapcell",
		"CockroachDB",
		"Aiven Valkey",
		"Cloudflare",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(outputStr, required) {
			t.Errorf("Expected output to contain '%s', but it was not found.\nOutput: %s", required, outputStr)
		}
	}

	// Verify deprecated providers are NOT mentioned
	deprecatedProviders := []string{
		"Railway",
		"Neon",
		"Back4app",
		"B4A",
	}

	for _, deprecated := range deprecatedProviders {
		if strings.Contains(strings.ToLower(outputStr), strings.ToLower(deprecated)) {
			t.Errorf("Output should not contain deprecated provider '%s', but it was found.\nOutput: %s", deprecated, outputStr)
		}
	}

	// Verify dry-run mode is indicated
	if !strings.Contains(outputStr, "dry-run") && !strings.Contains(outputStr, "Dry-run") {
		t.Errorf("Expected output to indicate dry-run mode.\nOutput: %s", outputStr)
	}
}

// Test_Deploy_Check_OpinionatedStack verifies that the deploy --check command
// validates the Opinionated Stack configuration
func Test_Deploy_Check_OpinionatedStack(t *testing.T) {
	os.Setenv("LOG_FORMAT", "off")
	defer os.Unsetenv("LOG_FORMAT")

	cmd := exec.Command("go", "run", "./cmd/gforge", "deploy", "--check")
	cmd.Dir = ".."
	cmd.Env = append(os.Environ(), "LOG_FORMAT=off")

	output, err := cmd.CombinedOutput()
	// Command may fail if prerequisites are missing, which is expected
	// We're just checking that it runs and mentions the right stack

	outputStr := string(output)

	// Verify it checks for Opinionated Stack components
	if !strings.Contains(outputStr, "CockroachDB") && !strings.Contains(outputStr, "COCKROACH") {
		t.Errorf("Expected preflight check to mention CockroachDB.\nOutput: %s\nError: %v", outputStr, err)
	}

	// Verify deprecated providers are NOT checked
	deprecatedProviders := []string{
		"Railway",
		"Neon",
		"Back4app",
	}

	for _, deprecated := range deprecatedProviders {
		if strings.Contains(strings.ToLower(outputStr), strings.ToLower(deprecated)) {
			t.Errorf("Preflight check should not mention deprecated provider '%s'.\nOutput: %s", deprecated, outputStr)
		}
	}
}

// Test_Deploy_NoDeprecatedProviders verifies that the deploy command code
// does not reference deprecated providers
func Test_Deploy_NoDeprecatedProviders(t *testing.T) {
	// This test searches the deploy.go file for deprecated provider references
	deployFile := "../cmd/gforge/cmd/deploy.go"
	content, err := os.ReadFile(deployFile)
	if err != nil {
		t.Fatalf("Failed to read deploy.go: %v", err)
	}

	fileContent := string(content)
	lowerContent := strings.ToLower(fileContent)

	// Check for deprecated provider references
	deprecatedTerms := map[string]string{
		"railway":  "Railway provider should be removed",
		"neon":     "Neon provider should be removed (excluding 'none')",
		"back4app": "Back4app provider should be removed",
		"b4a":      "B4A references should be removed",
	}

	for term, message := range deprecatedTerms {
		// Special handling for "neon" to avoid false positives with "none"
		if term == "neon" {
			if strings.Contains(lowerContent, "neon") && !isOnlyNoneReferences(fileContent) {
				t.Errorf("%s: Found '%s' in deploy.go", message, term)
			}
		} else {
			if strings.Contains(lowerContent, term) {
				t.Errorf("%s: Found '%s' in deploy.go", message, term)
			}
		}
	}
}

// Helper function to check if "neon" references are only "none"
func isOnlyNoneReferences(content string) bool {
	// If we find "neon" but it's always part of "none", return true
	lines := strings.Split(strings.ToLower(content), "\n")
	for _, line := range lines {
		if strings.Contains(line, "neon") && !strings.Contains(line, "none") {
			return false
		}
	}
	return true
}
