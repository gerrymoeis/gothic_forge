package main

import (
	"os"
	"testing"
	gfcmd "gothicforge3/cmd/gforge/cmd"
)

// Smoke test: running CLI with --help should not error or exit non-zero
func TestGforgeMain_Help(t *testing.T) {
	old := os.Args
	t.Cleanup(func() { os.Args = old })
	os.Args = []string{"gforge", "--help"}
	gfcmd.Execute()
}
