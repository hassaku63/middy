// Command fixtures writes the canonical JSON fixtures for the frontend.
//
//	go run ./cmd/fixtures -out ../frontend/src/shared/api/fixtures
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hassaku63/middy/backend/internal/fixtures"
)

func main() {
	out := flag.String("out", "", "directory to write fixtures into")
	flag.Parse()
	if *out == "" {
		fmt.Fprintln(os.Stderr, "fixtures: -out is required")
		os.Exit(2)
	}
	files, err := fixtures.Files()
	if err != nil {
		fmt.Fprintln(os.Stderr, "fixtures:", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(*out, 0o750); err != nil {
		fmt.Fprintln(os.Stderr, "fixtures:", err)
		os.Exit(1)
	}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(*out, name), data, 0o644); err != nil { //nolint:gosec // fixtures are public data
			fmt.Fprintln(os.Stderr, "fixtures:", err)
			os.Exit(1)
		}
		fmt.Println("wrote", filepath.Join(*out, name))
	}
}
