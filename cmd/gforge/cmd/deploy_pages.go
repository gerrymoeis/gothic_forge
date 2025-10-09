package cmd

import (
  "archive/zip"
  "context"
  "encoding/json"
  "fmt"
  "io"
  "net/http"
  "os"
  "os/exec"
  "path/filepath"
  "runtime"
  "strings"

  "github.com/spf13/cobra"
  "gothicforge3/internal/execx"
)

var (
  pagesOutDir    string
  pagesProject   string
  pagesDeployRun bool
)

var deployPagesCmd = &cobra.Command{
  Use:   "pages",
  Short: "Deploy static export to Cloudflare Pages (wrangler)",
  RunE: func(cmd *cobra.Command, args []string) error {
    banner()
    // 1) Export static site
    if pagesOutDir == "" { pagesOutDir = "dist" }
    exportOut = pagesOutDir
    if err := exportCmd.RunE(exportCmd, []string{}); err != nil {
      return err
    }

    // 2) Try wrangler CLI, else attempt install if requested, otherwise provide guidance
    if p, ok := execx.Look("wrangler"); ok {
      fmt.Println("wrangler found:", p)
      args := []string{"pages", "deploy", pagesOutDir}
      if strings.TrimSpace(pagesProject) != "" { args = append(args, "--project-name", pagesProject) }
      if pagesDeployRun {
        fmt.Println("Running:", "wrangler "+strings.Join(args, " "))
        ctx := context.Background()
        if strings.TrimSpace(pagesProject) != "" {
          if err := execx.RunInteractive(ctx, "wrangler pages deploy", "wrangler", "pages", "deploy", pagesOutDir, "--project-name", pagesProject); err != nil { return err }
        } else {
          if err := execx.RunInteractive(ctx, "wrangler pages deploy", "wrangler", "pages", "deploy", pagesOutDir); err != nil { return err }
        }
      } else {
        fmt.Println("Dry-run. To deploy with wrangler:")
        fmt.Println("  wrangler", strings.Join(args, " "))
      }
      return nil
    }

    // Attempt install if requested
    if deployInstall {
      if p2, err := ensureWranglerCLI(); err == nil {
        if p2 != "" { fmt.Println("wrangler installed:", p2) }
        // Re-run detection
        if p3, ok3 := execx.Look("wrangler"); ok3 {
          fmt.Println("wrangler found:", p3)
          args := []string{"pages", "deploy", pagesOutDir}
          if strings.TrimSpace(pagesProject) != "" { args = append(args, "--project-name", pagesProject) }
          if pagesDeployRun {
            fmt.Println("Running:", "wrangler "+strings.Join(args, " "))
            ctx := context.Background()
            if strings.TrimSpace(pagesProject) != "" {
              if err := execx.RunInteractive(ctx, "wrangler pages deploy", "wrangler", "pages", "deploy", pagesOutDir, "--project-name", pagesProject); err != nil { return err }
            } else {
              if err := execx.RunInteractive(ctx, "wrangler pages deploy", "wrangler", "pages", "deploy", pagesOutDir); err != nil { return err }
            }
          } else {
            fmt.Println("Dry-run. To deploy with wrangler:")
            fmt.Println("  wrangler", strings.Join(args, " "))
          }
          return nil
        }
      }
    }

    // Guidance when wrangler not installed (avoid npm; prefer brew or prebuilt binary)
    printWranglerInstallHelp()
    fmt.Println("Then run:")
    cmdLine := "wrangler pages deploy " + pagesOutDir
    if strings.TrimSpace(pagesProject) != "" { cmdLine += " --project-name " + pagesProject }
    fmt.Println("  "+cmdLine)
    return nil
  },
}

func init() {
  deployPagesCmd.Flags().StringVar(&pagesOutDir, "out", "dist", "export directory to deploy")
  deployPagesCmd.Flags().StringVar(&pagesProject, "project", "", "Cloudflare Pages project name")
  deployPagesCmd.Flags().BoolVar(&pagesDeployRun, "run", false, "execute wrangler if present (otherwise print instructions)")
  deployCmd.AddCommand(deployPagesCmd)
}

// printWranglerInstallHelp prints platform guidance to install wrangler without npm
func printWranglerInstallHelp() {
  fmt.Println("wrangler CLI not found in PATH.")
  fmt.Println("Install using one of:")
  switch runtime.GOOS {
  case "darwin":
    fmt.Println("  - Homebrew: brew install cloudflare/wrangler/wrangler")
    fmt.Println("  - Prebuilt binary: https://github.com/cloudflare/wrangler/releases")
  case "windows":
    fmt.Println("  - Prebuilt binary (ZIP): https://github.com/cloudflare/wrangler/releases")
    fmt.Println("  - Or use WSL with: bash <(curl -fsSL https://raw.githubusercontent.com/cloudflare/wrangler/refs/heads/master/scripts/install.sh)")
  default:
    fmt.Println("  - Prebuilt binary: https://github.com/cloudflare/wrangler/releases")
    fmt.Println("  - Or: bash <(curl -fsSL https://raw.githubusercontent.com/cloudflare/wrangler/refs/heads/master/scripts/install.sh)")
  }
  fmt.Println("Docs: https://developers.cloudflare.com/pages/framework-guides/deploy-a-static-site/")
}

// ensureWranglerCLI attempts to install wrangler using native package managers or direct binary.
// No npm fallback is used. Returns path to installed binary when possible.
func ensureWranglerCLI() (string, error) {
  // If already available
  if p, ok := execx.Look("wrangler"); ok { return p, nil }

  switch runtime.GOOS {
  case "darwin":
    if _, err := exec.LookPath("brew"); err == nil {
      if err := execx.Run(context.Background(), "brew install cloudflare/wrangler/wrangler", "brew", "install", "cloudflare/wrangler/wrangler"); err == nil {
        if p, ok := execx.Look("wrangler"); ok { return p, nil }
      }
    }
  case "windows":
    // Query GitHub Releases to find the correct asset per arch (x64/arm64),
    // handle both .exe and .zip assets.
    home, _ := os.UserHomeDir()
    binDir := filepath.Join(home, ".gforge", "bin")
    _ = os.MkdirAll(binDir, 0o755)
    archTag := "x64"
    if runtime.GOARCH == "arm64" { archTag = "arm64" }
    assetURL, isZip, err := getLatestWranglerAssetURL(archTag)
    if err == nil && assetURL != "" {
      tmp := filepath.Join(os.TempDir(), "wrangler-asset")
      if isZip { tmp += ".zip" } else { tmp += ".exe" }
      if err := downloadSimple(assetURL, tmp); err == nil {
        if isZip {
          exePath, uerr := unzipWranglerExe(tmp, binDir)
          if uerr == nil {
            _ = os.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
            return exePath, nil
          }
        } else {
          dest := filepath.Join(binDir, "wrangler.exe")
          // Move file into place
          // On Windows, rename across volumes may fail; fallback to copy
          if rerr := os.Rename(tmp, dest); rerr != nil {
            in, _ := os.Open(tmp)
            defer in.Close()
            out, _ := os.Create(dest)
            defer out.Close()
            _, _ = io.Copy(out, in)
            _ = os.Remove(tmp)
          }
          _ = os.Chmod(dest, 0o755)
          _ = os.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
          return dest, nil
        }
      }
    }
    // Fallback: try direct exe URL if API call or asset download failed (best-effort)
    {
      direct := "https://github.com/cloudflare/wrangler/releases/latest/download/wrangler.exe"
      dest := filepath.Join(binDir, "wrangler.exe")
      if err := downloadSimple(direct, dest); err == nil {
        _ = os.Chmod(dest, 0o755)
        _ = os.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
        return dest, nil
      }
    }
  default: // linux
    // Attempt official install script if curl+bash exist
    if _, berr := exec.LookPath("bash"); berr == nil {
      if _, cerr := exec.LookPath("curl"); cerr == nil {
        cmd := exec.Command("bash", "-lc", "bash <(curl -fsSL https://raw.githubusercontent.com/cloudflare/wrangler/refs/heads/master/scripts/install.sh)")
        cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
        _ = cmd.Run()
        if p, ok := execx.Look("wrangler"); ok { return p, nil }
      }
    }
  }
  return "", fmt.Errorf("wrangler auto-install not available on this platform; see help")
}

func downloadSimple(url, dest string) error {
  resp, err := http.Get(url)
  if err != nil { return err }
  defer resp.Body.Close()
  if resp.StatusCode >= 300 { return fmt.Errorf("download failed: %s", resp.Status) }
  f, err := os.Create(dest)
  if err != nil { return err }
  defer f.Close()
  _, err = io.Copy(f, resp.Body)
  return err
}

// getLatestWranglerAssetURL queries GitHub Releases for wrangler and returns an asset URL
// appropriate for Windows with the specified arch tag ("x64" or "arm64").
// It returns the URL and whether the asset is a zip file.
func getLatestWranglerAssetURL(archTag string) (string, bool, error) {
  req, err := http.NewRequest("GET", "https://api.github.com/repos/cloudflare/wrangler/releases/latest", nil)
  if err != nil { return "", false, err }
  req.Header.Set("User-Agent", "gforge-installer")
  resp, err := http.DefaultClient.Do(req)
  if err != nil { return "", false, err }
  defer resp.Body.Close()
  if resp.StatusCode >= 300 {
    b, _ := io.ReadAll(resp.Body)
    return "", false, fmt.Errorf("github api: %s: %s", resp.Status, string(b))
  }
  var data struct {
    Assets []struct {
      Name string `json:"name"`
      URL  string `json:"browser_download_url"`
    } `json:"assets"`
  }
  dec := json.NewDecoder(resp.Body)
  if err := dec.Decode(&data); err != nil { return "", false, err }
  // Prefer .zip, then .exe; require windows + archTag match
  // Common patterns: windows-x64.zip, windows-arm64.zip, wrangler.exe
  var exeURL, zipURL string
  for _, a := range data.Assets {
    n := strings.ToLower(a.Name)
    if !strings.Contains(n, "windows") && !strings.HasSuffix(n, ".exe") { continue }
    // arch match
    if archTag == "x64" {
      if !(strings.Contains(n, "x64") || strings.Contains(n, "amd64")) && strings.HasSuffix(n, ".exe") {
        // plain wrangler.exe (no arch in name) – accept as fallback
        exeURL = a.URL
        continue
      }
    } else if archTag == "arm64" {
      if !strings.Contains(n, "arm64") { continue }
    }
    if strings.HasSuffix(n, ".zip") { zipURL = a.URL }
    if strings.HasSuffix(n, ".exe") { exeURL = a.URL }
  }
  if zipURL != "" { return zipURL, true, nil }
  if exeURL != "" { return exeURL, false, nil }
  return "", false, fmt.Errorf("no suitable wrangler asset found for windows-%s", archTag)
}

// unzipWranglerExe extracts wrangler.exe from a zip archive to destDir and returns the path.
func unzipWranglerExe(zipPath, destDir string) (string, error) {
  zr, err := zip.OpenReader(zipPath)
  if err != nil { return "", err }
  defer zr.Close()
  var exeFile *zip.File
  for _, f := range zr.File {
    name := strings.ToLower(f.Name)
    if strings.HasSuffix(name, "wrangler.exe") || strings.HasSuffix(name, "/wrangler.exe") {
      exeFile = f
      break
    }
  }
  if exeFile == nil { return "", fmt.Errorf("wrangler.exe not found in zip") }
  rc, err := exeFile.Open()
  if err != nil { return "", err }
  defer rc.Close()
  dest := filepath.Join(destDir, "wrangler.exe")
  out, err := os.Create(dest)
  if err != nil { return "", err }
  defer out.Close()
  if _, err := io.Copy(out, rc); err != nil { return "", err }
  _ = os.Chmod(dest, 0o755)
  return dest, nil
}
