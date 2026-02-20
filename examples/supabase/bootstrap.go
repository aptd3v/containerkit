// Package main: write embedded Supabase volume files to disk for the example.
// All required config files are embedded so the example runs offline with no GitHub fetch.
package main

import (
	"embed"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
)

//go:embed volumes_embed
var volumesEmbed embed.FS

// requiredVolumeFiles are bind-mounted into containers; each must be a file, not a directory.
var requiredVolumeFiles = []string{
	"api/kong.yml", "db/realtime.sql", "db/webhooks.sql", "db/roles.sql", "db/jwt.sql",
	"db/_supabase.sql", "db/logs.sql", "db/pooler.sql", "logs/vector.yml",
	"pooler/pooler.exs", "functions/main/index.ts",
}

// embedPathFor maps a required volume path to its embed path (Go embed excludes names starting with _).
func embedPathFor(rel string) string {
	if rel == "db/_supabase.sql" {
		return "db/supabase.sql"
	}
	return rel
}

// bootstrapSupabaseVolumes writes embedded volume files under basePath and creates
// required directories (api, db, logs, pooler, storage, snippets, functions, db/data).
// Ensures every required file is present as a file (removes any existing directory at that path).
func bootstrapSupabaseVolumes(basePath string) {
	dirs := []string{"api", "db", "logs", "pooler", "storage", "snippets", "functions"}
	for _, d := range dirs {
		p := filepath.Join(basePath, d)
		if err := os.MkdirAll(p, 0755); err != nil {
			log.Fatalf("bootstrap volumes: mkdir %s: %v", p, err)
		}
	}
	if err := os.MkdirAll(filepath.Join(basePath, "db", "data"), 0755); err != nil {
		log.Fatalf("bootstrap volumes: mkdir db/data: %v", err)
	}

	root := "volumes_embed"
	err := fs.WalkDir(volumesEmbed, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		writeRel := rel
		if rel == "db/supabase.sql" {
			writeRel = "db/_supabase.sql" // container bind-mount expects _supabase.sql; embed excludes leading _
		}
		data, err := fs.ReadFile(volumesEmbed, path)
		if err != nil {
			return err
		}
		writeVolumeFile(basePath, writeRel, data)
		return nil
	})
	if err != nil {
		log.Fatalf("bootstrap volumes: %v", err)
	}
	// Guarantee every bind-mount target is a file (fix leftover directories from old runs).
	for _, rel := range requiredVolumeFiles {
		fullPath := filepath.Join(basePath, rel)
		info, err := os.Stat(fullPath)
		if err == nil && info.Mode().IsRegular() && info.Size() > 0 {
			continue // already a non-empty file
		}
		replacedDir := false
		if err == nil && info.IsDir() {
			if err := os.RemoveAll(fullPath); err != nil {
				log.Fatalf("bootstrap volumes: remove dir %s: %v", fullPath, err)
			}
			log.Printf("bootstrapped %s (replaced directory)", rel)
			replacedDir = true
		}
		embedPath := path.Join(root, embedPathFor(rel)) // embed.FS uses forward slashes; _supabase.sql -> supabase.sql
		data, err := fs.ReadFile(volumesEmbed, embedPath)
		if err != nil {
			log.Fatalf("bootstrap volumes: read embedded %s: %v", rel, err)
		}
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			log.Fatalf("bootstrap volumes: mkdir for %s: %v", rel, err)
		}
		if err := os.WriteFile(fullPath, data, 0644); err != nil {
			log.Fatalf("bootstrap volumes: write %s: %v", fullPath, err)
		}
		if !replacedDir {
			log.Printf("bootstrapped %s", rel)
		}
	}
}

func writeVolumeFile(basePath, rel string, data []byte) {
	fullPath := filepath.Join(basePath, rel)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		log.Fatalf("bootstrap volumes: mkdir for %s: %v", rel, err)
	}
	if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
		if err := os.RemoveAll(fullPath); err != nil {
			log.Fatalf("bootstrap volumes: remove dir %s: %v", fullPath, err)
		}
	}
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		log.Fatalf("bootstrap volumes: write %s: %v", fullPath, err)
	}
	log.Printf("bootstrapped %s", rel)
}
