package cmd

import (
    "fmt"
    "io"
    "io/fs"
    "os"
    "path/filepath"
    "strings"

    "github.com/fatih/color"
    "github.com/spf13/cobra"

    "gothicforge/app/ssg"
    "gothicforge/internal/execx"
)

var (
    outDir string
)

var exportCmd = &cobra.Command{
    Use:   "export",
    Short: "Export static pages (SSG) to an output directory",
    RunE: func(cmd *cobra.Command, args []string) error {
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

        // 1) Export registered SSG pages
        pages := ssg.Pages()
        for _, p := range pages {
            html, err := p.Render(cmd.Context())
            if err != nil {
                return fmt.Errorf("render %s: %w", p.Path, err)
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

        // 2) Copy /static assets
        if err := copyDir("./app/static", filepath.Join(outDir, "static")); err != nil {
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
    }
    return filepath.Join(out, clean, "index.html")
}

func copyDir(src, dst string) error {
    // ensure dst exists
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
        // copy file
        in, err := os.Open(path)
        if err != nil {
            return err
        }
        defer in.Close()
        out, err := os.Create(target)
        if err != nil {
            return err
        }
        if _, err := io.Copy(out, in); err != nil {
            out.Close()
            return err
        }
        return out.Close()
    })
}

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
