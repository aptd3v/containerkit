// Package main: write embedded Astronomy Shop bind-mount assets under a volumes path.
package main

import (
	"embed"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

//go:embed embed
var assetsEmbed embed.FS

// bootstrapVolumes materializes embedded flagd, postgres, collector, jaeger,
// grafana, and prometheus configs under basePath (no network fetch).
// Directories and files are made 0777/0666 so containers with arbitrary UIDs
// (common on Linux) can always access the bind mounts.
func bootstrapVolumes(basePath string) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		log.Fatalf("bootstrap volumes: mkdir %s: %v", basePath, err)
	}
	root := "embed"
	err := fs.WalkDir(assetsEmbed, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		dest := filepath.Join(basePath, rel)
		if d.IsDir() {
			return os.MkdirAll(dest, 0755)
		}
		data, err := fs.ReadFile(assetsEmbed, path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0644)
	})
	if err != nil {
		log.Fatalf("bootstrap volumes: %v", err)
	}

	// Make world-accessible so containers (which may run as different UIDs on Linux)
	// can read/write bind mounts regardless of host user. Harmless on macOS.
	_ = os.Chmod(basePath, 0777)
	_ = fs.WalkDir(os.DirFS(basePath), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // best effort
		}
		dest := filepath.Join(basePath, p)
		if d.IsDir() {
			_ = os.Chmod(dest, 0777)
		} else {
			_ = os.Chmod(dest, 0666)
		}
		return nil
	})
}
