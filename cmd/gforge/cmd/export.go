package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gothicforge3/app/routes"
	"gothicforge3/internal/execx"
	"gothicforge3/internal/server"
)

var (
	exportOut string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export static HTML to a directory (SSG)",
	RunE: func(cmd *cobra.Command, args []string) error {
		banner()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		// Ensure SEO files (sitemap.xml, robots.txt) exist prior to copy
		if err := writeSEOFiles(); err != nil {
			fmt.Printf("seo files generation warning: %v\n", err)
		}

		// Ensure Templ + CSS are fresh unless skipped for tests/CI
		if os.Getenv("GFORGE_SKIP_TOOLS") == "" {
			if templPath, err := ensureTool("templ", "github.com/a-h/templ/cmd/templ@latest"); err == nil {
				_ = execx.Run(ctx, "templ", templPath, "generate", "-include-version=false", "-include-timestamp=false")
			}
			if gwPath, err := ensureTool("gotailwindcss", "github.com/gotailwindcss/tailwind/cmd/gotailwindcss@latest"); err == nil {
				_ = execx.Run(ctx, "gotailwindcss build", gwPath, "build", "-o", "./app/styles/output.css", "./app/styles/tailwind.input.css")
			}
		}

		outDir := exportOut
		if outDir == "" { outDir = "dist" }
		if err := os.MkdirAll(outDir, 0o755); err != nil { return err }

		r := server.New()
		routes.Register(r)

		// Derive URLs: from sitemap helper, then convert to paths
		base := strings.TrimSpace(os.Getenv("SITE_BASE_URL"))
		if base == "" { base = "/" }
		urlList := collectSitemapURLs(base)
		paths := toPaths(urlList, base)

		// Render each path to an index.html
		for _, p := range paths {
			req := httptest.NewRequest(http.MethodGet, p, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK { continue }
			targetDir := filepath.Join(outDir, strings.TrimPrefix(p, "/"))
			if p == "/" {
				targetDir = outDir
			}
			if err := os.MkdirAll(targetDir, 0o755); err != nil { return err }
			file := filepath.Join(targetDir, "index.html")
			if err := os.WriteFile(file, rec.Body.Bytes(), 0o644); err != nil { return err }
		}
		// Copy assets: app/static -> dist/static; app/styles -> dist/static/styles
		if err := copyDir("app/static", filepath.Join(outDir, "static")); err != nil { return err }
		if err := copyDir("app/styles", filepath.Join(outDir, "static", "styles")); err != nil { return err }

		fmt.Println("Static export complete:", outDir)
		return nil
	},
}

func init() {
	exportCmd.Flags().StringVarP(&exportOut, "out", "o", "dist", "output directory")
	rootCmd.AddCommand(exportCmd)
}

// toPaths transforms absolute URLs into path-only list anchored at base.
func toPaths(urls []string, base string) []string {
	b := strings.TrimRight(base, "/")
	res := make([]string, 0, len(urls))
	for _, u := range urls {
		s := strings.TrimSpace(u)
		if s == "" { continue }
		if s == "/" { res = append(res, "/"); continue }
		if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
			if b != "" && b != "/" && strings.HasPrefix(s, b) {
				p := s[len(b):]
				if p == "" { p = "/" }
				res = append(res, p)
			}
			continue
		}
		if !strings.HasPrefix(s, "/") { s = "/" + s }
		res = append(res, s)
	}
	return res
}

// copyDir copies src directory recursively into dst.
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil { return err }
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := os.ReadFile(path)
		if err != nil { return err }
		return os.WriteFile(target, b, 0o644)
	})
}
