package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/a-h/templ"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"gothicforge/app/templates"
)

var (
	outDir string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export static pages (SSG) to an output directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		if outDir == "" {
			outDir = "dist"
		}
		color.Cyan("Exporting static pages to %s...", outDir)
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return fmt.Errorf("create out dir: %w", err)
		}

		// 1) Export pages
		pages := []struct {
			Path   string
			Render func(context.Context) (string, error)
		}{
			{
				Path: "/",
				Render: func(ctx context.Context) (string, error) {
					return renderToString(ctx, templates.Index())
				},
			},
			{
				Path: "/counter",
				Render: func(ctx context.Context) (string, error) {
					return renderToString(ctx, templates.CounterPage(0))
				},
			},
		}

		for _, p := range pages {
			html, err := p.Render(context.Background())
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

		color.HiGreen("Done.")
		return nil
	},
}

func init() {
	exportCmd.Flags().StringVarP(&outDir, "out", "o", "dist", "output directory for static export")
	rootCmd.AddCommand(exportCmd)
}

func renderToString(ctx context.Context, c templ.Component) (string, error) {
	var buf bytes.Buffer
	if err := c.Render(ctx, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func pathFor(out, routePath string) string {
	// map "/" -> out/index.html, "/about" -> out/about/index.html
	clean := strings.TrimPrefix(routePath, "/")
	if clean == "" {
		return filepath.Join(out, "index.html")
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
