package cmd

import (
  "context"
  "fmt"
  "os"
  "path/filepath"
  "sort"
  "strings"
  "time"

  "gothicforge3/internal/execx"

  "github.com/spf13/cobra"
)

var (
  buildOS   string
  buildArch string
)

var buildCmd = &cobra.Command{
  Use:   "build",
  Short: "Build production binary and assets",
  RunE: func(cmd *cobra.Command, args []string) error {
    banner()
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
    defer cancel()

    // Ensure tools
    if templPath, err := ensureTool("templ", "github.com/a-h/templ/cmd/templ@latest"); err == nil {
      _ = execx.Run(ctx, "templ", templPath, "generate", "-include-version=false", "-include-timestamp=false")
    } else {
      fmt.Printf("templ not available: %v\n", err)
    }
    if gwPath, err := ensureTool("gotailwindcss", "github.com/gotailwindcss/tailwind/cmd/gotailwindcss@latest"); err == nil {
      _ = execx.Run(ctx, "gotailwindcss build", gwPath, "build", "-o", "./app/styles/output.css", "./app/styles/tailwind.input.css")
    } else {
      fmt.Printf("gotailwindcss not available: %v\n", err)
    }

    // Build server
    _ = os.MkdirAll("bin", 0o755)
    out := filepath.Join("bin", "server")
    if buildOS != "" && buildArch != "" {
      os.Setenv("GOOS", buildOS)
      os.Setenv("GOARCH", buildArch)
    }
    fmt.Println("Building server ->", out)
    if err := execx.Run(ctx, "go build", "go", "build", "-o", out, "./cmd/server"); err != nil {
      return err
    }

    // SEO files (auto-generate sitemap.xml and robots.txt)
    if err := writeSEOFiles(); err != nil {
      fmt.Printf("seo files generation warning: %v\n", err)
    }

    fmt.Println("────────────────────────────────────────")
    fmt.Println("Build complete.")
    return nil
  },
}

func init() {
  buildCmd.Flags().StringVar(&buildOS, "os", "", "target OS (optional)")
  buildCmd.Flags().StringVar(&buildArch, "arch", "", "target arch (optional)")
  rootCmd.AddCommand(buildCmd)
}

// writeSEOFiles writes sitemap.xml and robots.txt into app/static/.
// It uses SITE_BASE_URL if provided; otherwise falls back to "/".
func writeSEOFiles() error {
  base := strings.TrimSpace(os.Getenv("SITE_BASE_URL"))
  if base == "" { base = "/" }
  // normalize base (no trailing slash unless root)
  if base != "/" { base = strings.TrimRight(base, "/") }
  // respect GFORGE_BASEDIR when set (useful for tests)
  baseDir := strings.TrimSpace(os.Getenv("GFORGE_BASEDIR"))
  var staticDir string
  if baseDir != "" {
    staticDir = filepath.Join(baseDir, "app", "static")
  } else {
    staticDir = filepath.Join("app", "static")
  }
  if err := os.MkdirAll(staticDir, 0o755); err != nil { return err }

  // Collect URLs: root + registry from app/sitemap/urls.txt
  urls := collectSitemapURLs(base)
  // Deduplicate and stable sort
  uniq := map[string]struct{}{}
  out := make([]string, 0, len(urls))
  for _, u := range urls { if _, ok := uniq[u]; !ok { uniq[u] = struct{}{}; out = append(out, u) } }
  sort.Strings(out)

  // sitemap.xml
  siteMapPath := filepath.Join(staticDir, "sitemap.xml")
  sb := &strings.Builder{}
  sb.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
  sb.WriteString("<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n")
  for _, u := range out {
    sb.WriteString("  <url><loc>")
    sb.WriteString(u)
    sb.WriteString("</loc></url>\n")
  }
  sb.WriteString("</urlset>\n")
  if err := os.WriteFile(siteMapPath, []byte(sb.String()), 0o644); err != nil { return err }

  // robots.txt
  robotsPath := filepath.Join(staticDir, "robots.txt")
  robots := &strings.Builder{}
  robots.WriteString("User-agent: *\n")
  robots.WriteString("Allow: /\n")
  robots.WriteString("Sitemap: ")
  if base == "/" {
    robots.WriteString("/sitemap.xml\n")
  } else {
    robots.WriteString(base + "/sitemap.xml\n")
  }
  if err := os.WriteFile(robotsPath, []byte(robots.String()), 0o644); err != nil { return err }
  return nil
}

// collectSitemapURLs reads app/sitemap/urls.txt if present and returns a list of absolute URLs.
// Lines starting with http(s) are used as-is; other lines are treated as paths relative to base.
// Blank lines and lines starting with '#' are ignored.
func collectSitemapURLs(base string) []string {
  urls := []string{}
  // Always include root
  root := base
  if root == "/" { root = "/" }
  urls = append(urls, root)
  // Extras
  baseDir := strings.TrimSpace(os.Getenv("GFORGE_BASEDIR"))
  var extra string
  if baseDir != "" {
    extra = filepath.Join(baseDir, "app", "sitemap", "urls.txt")
  } else {
    extra = filepath.Join("app", "sitemap", "urls.txt")
  }
  if b, err := os.ReadFile(extra); err == nil {
    lines := strings.Split(string(b), "\n")
    for _, ln := range lines {
      t := strings.TrimSpace(ln)
      if t == "" || strings.HasPrefix(t, "#") { continue }
      if strings.HasPrefix(t, "http://") || strings.HasPrefix(t, "https://") {
        urls = append(urls, strings.TrimRight(t, "/"))
        continue
      }
      // path
      if !strings.HasPrefix(t, "/") { t = "/" + t }
      if base == "/" { urls = append(urls, t) } else { urls = append(urls, base + t) }
    }
  }
  return urls
}
