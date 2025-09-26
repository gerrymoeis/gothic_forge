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
)

var (
	outDir string
)

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
