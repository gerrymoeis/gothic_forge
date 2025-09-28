package cmd

import (
    "bytes"
    "fmt"
    "image"
    "image/jpeg"
    "image/png"
    "io"
    "io/fs"
    "os"
    "os/exec"
    "path/filepath"
    "strconv"
    "strings"

    "github.com/fatih/color"
    "github.com/spf13/cobra"
    "github.com/tdewolff/minify/v2"
    "github.com/tdewolff/minify/v2/css"
    "github.com/tdewolff/minify/v2/html"
    "github.com/tdewolff/minify/v2/js"
    "github.com/tdewolff/minify/v2/svg"

    "gothicforge/app/ssg"
    "gothicforge/app/templates"
    "gothicforge/internal/execx"
)

var (
    outDir string
)

var exportCmd = &cobra.Command{
    Use:   "export",
    Short: "Export static pages (SSG) to an output directory",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Minifier for production assets
        m := minify.New()
        m.Add("text/css", &css.Minifier{})
        m.Add("text/javascript", &js.Minifier{})
        m.Add("image/svg+xml", &svg.Minifier{})
        m.Add("text/html", &html.Minifier{})

        // Ensure templ code is generated once before export so SSG is up-to-date.
        if templPath, err := ensureTool("templ", "github.com/a-h/templ/cmd/templ@latest"); err == nil {
            _ = execx.Run(cmd.Context(), "templ generate", templPath, "generate", "-include-version=false", "-include-timestamp=false")
        } else {
            color.Yellow("templ not available and auto-install failed: %v", err)
        }
        if outDir == "" {
            outDir = "dist"
        }
        color.Cyan("Exporting static pages to %s...", outDir)
        if err := os.MkdirAll(outDir, 0o755); err != nil {
            return fmt.Errorf("create out dir: %w", err)
        }

        // 1) Export registered SSG pages (minified HTML)
        pages := ssg.Pages()
        for _, p := range pages {
            html, err := p.Render(cmd.Context())
            if err != nil {
                return fmt.Errorf("render %s: %w", p.Path, err)
            }
            // Minify HTML
            if minified, merr := m.String("text/html", html); merr == nil {
                html = minified
            }
            target := pathFor(outDir, p.Path)
            if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
                return fmt.Errorf("mkdir for %s: %w", target, err)
            }
            if err := os.WriteFile(target, []byte(html), 0o644); err != nil {
                return fmt.Errorf("write %s: %w", target, err)
            }
            color.Green("✔ page %s -> %s", p.Path, target)
        }

        // 2) Copy /static assets (minify css/js/svg)
        if err := copyDirMinified(m, "./app/static", filepath.Join(outDir, "static")); err != nil {
            return fmt.Errorf("copy static: %w", err)
        }
        color.Green("✔ static -> %s", filepath.Join(outDir, "static"))

        // 2a) Also place robots.txt and sitemap.xml at the root if present under static
        for _, name := range []string{"robots.txt", "sitemap.xml"} {
            src := filepath.Join(outDir, "static", name)
            dst := filepath.Join(outDir, name)
            if _, err := os.Stat(src); err == nil {
                if err := copyFile(src, dst); err != nil {
                    return fmt.Errorf("copy %s to root: %w", name, err)
                }
                color.Green("✔ %s -> %s", name, dst)
            }
        }

        // 2b) If robots.txt or sitemap.xml are missing, generate sensible defaults at root
        robotsRoot := filepath.Join(outDir, "robots.txt")
        robotsStatic := filepath.Join(outDir, "static", "robots.txt")
        if _, err := os.Stat(robotsRoot); err != nil {
            if _, err2 := os.Stat(robotsStatic); err2 != nil {
                if err := os.WriteFile(robotsRoot, []byte(generateRobots()), 0o644); err != nil {
                    return fmt.Errorf("write robots.txt: %w", err)
                }
                color.Green("✔ robots.txt -> %s", robotsRoot)
            }
        }
        sitemapRoot := filepath.Join(outDir, "sitemap.xml")
        sitemapStatic := filepath.Join(outDir, "static", "sitemap.xml")
        if _, err := os.Stat(sitemapRoot); err != nil {
            if _, err2 := os.Stat(sitemapStatic); err2 != nil {
                if err := os.WriteFile(sitemapRoot, []byte(generateSitemap()), 0o644); err != nil {
                    return fmt.Errorf("write sitemap.xml: %w", err)
                }
                color.Green("✔ sitemap.xml -> %s", sitemapRoot)
            }
        }

        // 3) Export a root-level 404.html for static hosts (minimal HTML)
        notFoundHTML := "<!doctype html><html><head><meta charset=\"utf-8\"><title>404</title></head><body><h1>404</h1><p>Not Found</p></body></html>"
        if err := os.WriteFile(filepath.Join(outDir, "404.html"), []byte(notFoundHTML), 0o644); err != nil {
            return fmt.Errorf("write 404.html: %w", err)
        }
        color.Green("✔ 404 -> %s", filepath.Join(outDir, "404.html"))

        color.HiGreen("Done.")
        return nil
    },
}

func init() {
    exportCmd.Flags().StringVarP(&outDir, "out", "o", "dist", "output directory for static export")
    rootCmd.AddCommand(exportCmd)
}
func pathFor(out, routePath string) string {
    // map "/" -> out/index.html, "/about" -> out/about/index.html
    clean := strings.TrimPrefix(routePath, "/")
    if clean == "" {
        return filepath.Join(out, "index.html")
    }
    return filepath.Join(out, clean, "index.html")
}

// removed unused copyDir; using copyDirMinified and copyFile instead

// copyFile copies a single file from src to dst, creating parent directories for dst.
func copyFile(src, dst string) error {
    in, err := os.Open(src)
    if err != nil {
        return err
    }
    defer in.Close()
    if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
        return err
    }
    out, err := os.Create(dst)
    if err != nil {
        return err
    }
    if _, err := io.Copy(out, in); err != nil {
        out.Close()
        return err
    }
    return out.Close()
}

// copyDirMinified copies files from src to dst and minifies CSS/JS/SVG assets.
func copyDirMinified(m *minify.M, src, dst string) error {
    if err := os.MkdirAll(dst, 0o755); err != nil {
        return err
    }
    return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        rel, err := filepath.Rel(src, path)
        if err != nil {
            return err
        }
        target := filepath.Join(dst, rel)
        if d.IsDir() {
            return os.MkdirAll(target, 0o755)
        }
        ext := strings.ToLower(filepath.Ext(path))
        switch ext {
        case ".css":
            b, err := os.ReadFile(path)
            if err != nil { return err }
            out, err := m.Bytes("text/css", b)
            if err != nil { out = b }
            return os.WriteFile(target, out, 0o644)
        case ".js":
            b, err := os.ReadFile(path)
            if err != nil { return err }
            out, err := m.Bytes("text/javascript", b)
            if err != nil { out = b }
            return os.WriteFile(target, out, 0o644)
        case ".svg":
            b, err := os.ReadFile(path)
            if err != nil { return err }
            out, err := m.Bytes("image/svg+xml", b)
            if err != nil { out = b }
            return os.WriteFile(target, out, 0o644)
        case ".png":
            if err := optimizePNG(path, target); err == nil { return nil }
            return copyFile(path, target)
        case ".jpg", ".jpeg":
            if err := optimizeJPEG(path, target); err == nil { return nil }
            return copyFile(path, target)
        case ".gif":
            if err := optimizeGIF(path, target); err == nil { return nil }
            return copyFile(path, target)
        case ".mp4", ".mov":
            if err := optimizeMP4MOV(path, target); err == nil { return nil }
            return copyFile(path, target)
        case ".webm", ".mp3", ".wav", ".ogg", ".m4a":
            if err := optimizeAudioGeneric(path, target); err == nil { return nil }
            return copyFile(path, target)
        default:
            // copy as-is
            in, err := os.Open(path)
            if err != nil { return err }
            defer in.Close()
            if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil { return err }
            out, err := os.Create(target)
            if err != nil { return err }
            if _, err := io.Copy(out, in); err != nil { out.Close(); return err }
            return out.Close()
        }
    })
}

// ---- Media optimization helpers ----
func hasTool(name string) bool {
    _, err := exec.LookPath(name)
    return err == nil
}

func envLossy() bool {
    v := strings.TrimSpace(os.Getenv("GFORGE_MEDIA_LOSSY"))
    return v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes")
}

func jpegQuality() int {
    if s := strings.TrimSpace(os.Getenv("GFORGE_JPEG_QUALITY")); s != "" {
        if n, err := strconv.Atoi(s); err == nil && n >= 1 && n <= 100 {
            return n
        }
    }
    return 85
}

func optimizePNG(src, dst string) error {
    // Prefer oxipng for lossless optimization.
    if hasTool("oxipng") {
        if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil { return err }
        cmd := exec.Command("oxipng", "-o", "3", "--strip", "all", "--preserve", "--quiet", "--out", dst, src)
        return cmd.Run()
    }
    // Optional lossy fallback: re-encode PNG in Go (often minimal gains). Only when GFORGE_MEDIA_LOSSY=1.
    if envLossy() {
        f, err := os.Open(src)
        if err != nil { return err }
        defer f.Close()
        img, _, err := image.Decode(f)
        if err != nil { return err }
        if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil { return err }
        // Go's PNG is lossless; set BestCompression to try shrinking a bit.
        var buf bytes.Buffer
        enc := png.Encoder{CompressionLevel: png.BestCompression}
        if err := enc.Encode(&buf, img); err != nil { return err }
        return os.WriteFile(dst, buf.Bytes(), 0o644)
    }
    return fmt.Errorf("no optimizer available")
}

func optimizeJPEG(src, dst string) error {
    // Prefer jpegoptim for lossless metadata strip + progressive.
    if hasTool("jpegoptim") {
        if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil { return err }
        // jpegoptim writes in-place or to --dest directory.
        cmd := exec.Command("jpegoptim", "--strip-all", "--all-progressive", "--quiet", "--dest="+filepath.Dir(dst), src)
        if err := cmd.Run(); err == nil {
            return nil
        }
    }
    // Lossy fallback with Go encoder if enabled.
    if envLossy() {
        f, err := os.Open(src)
        if err != nil { return err }
        defer f.Close()
        img, _, err := image.Decode(f)
        if err != nil { return err }
        if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil { return err }
        var buf bytes.Buffer
        if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: jpegQuality()}); err != nil { return err }
        return os.WriteFile(dst, buf.Bytes(), 0o644)
    }
    return fmt.Errorf("no optimizer available")
}

func optimizeGIF(src, dst string) error {
    if hasTool("gifsicle") {
        if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil { return err }
        cmd := exec.Command("gifsicle", "-O3", "-o", dst, src)
        return cmd.Run()
    }
    return fmt.Errorf("no optimizer available")
}

func optimizeMP4MOV(src, dst string) error {
    if hasTool("ffmpeg") {
        if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil { return err }
        // Lossless container optimization: faststart for MP4/MOV, stream copy.
        cmd := exec.Command("ffmpeg", "-y", "-i", src, "-c", "copy", "-movflags", "+faststart", dst)
        return cmd.Run()
    }
    return fmt.Errorf("no optimizer available")
}

func optimizeAudioGeneric(src, dst string) error {
    if hasTool("ffmpeg") {
        if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil { return err }
        // Strip metadata, stream copy (no quality change).
        cmd := exec.Command("ffmpeg", "-y", "-i", src, "-map_metadata", "-1", "-c", "copy", dst)
        return cmd.Run()
    }
    return fmt.Errorf("no optimizer available")
}
// generateRobots returns a minimal robots.txt with an optional Sitemap line.
func generateRobots() string {
    var b strings.Builder
    b.WriteString("User-agent: *\n")
    b.WriteString("Allow: /\n")
    sm := templates.ResolveCanonical("/sitemap.xml")
    if sm != "" {
        b.WriteString("Sitemap: ")
        b.WriteString(sm)
        b.WriteString("\n")
    }
    return b.String()
}

// generateSitemap renders a minimal sitemap.xml using registered SSG pages.
func generateSitemap() string {
    var b strings.Builder
    b.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
    b.WriteString("<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n")
    seen := make(map[string]struct{})
    add := func(p string) {
        if p == "" {
            return
        }
        if _, ok := seen[p]; ok {
            return
        }
        seen[p] = struct{}{}
        loc := templates.ResolveCanonical(p)
        if loc == "" {
            loc = p
        }
        b.WriteString("  <url><loc>")
        b.WriteString(loc)
        b.WriteString("</loc></url>\n")
    }
    add("/")
    for _, sp := range ssg.Pages() {
        p := strings.TrimSpace(sp.Path)
        if p == "" {
            continue
        }
        add(p)
    }
    b.WriteString("</urlset>\n")
    return b.String()
}
