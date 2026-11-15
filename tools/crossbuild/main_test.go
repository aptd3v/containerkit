package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestTargets(t *testing.T) {
	t.Parallel()
	want := map[string]bool{
		"linux/amd64":   false,
		"linux/arm64":   false,
		"darwin/amd64":  false,
		"darwin/arm64":  false,
		"windows/amd64": false,
		"windows/arm64": false,
	}
	if len(targets) != len(want) {
		t.Fatalf("len(targets) = %d", len(targets))
	}
	for _, t0 := range targets {
		key := t0.goos + "/" + t0.goarch
		if _, ok := want[key]; !ok {
			t.Fatalf("unexpected %s", key)
		}
		want[key] = true
	}
	for key, seen := range want {
		if !seen {
			t.Fatalf("missing %s", key)
		}
	}
}

func TestArchiveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "bin")
	if err := os.WriteFile(src, []byte("containerkit"), 0o755); err != nil {
		t.Fatal(err)
	}
	tarPath := filepath.Join(dir, "bin.tar.gz")
	if err := tarball(tarPath, "containerkit", src); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(tarPath)
	if err != nil {
		t.Fatal(err)
	}
	gz, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	tr := tar.NewReader(gz)
	hdr, err := tr.Next()
	if err != nil {
		t.Fatal(err)
	}
	if hdr.Name != "containerkit" {
		t.Fatalf("tar name %q", hdr.Name)
	}
	body, err := io.ReadAll(tr)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "containerkit" {
		t.Fatalf("tar body %q", body)
	}
	if _, err := tr.Next(); err != io.EOF {
		t.Fatalf("tar trailer: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}

	zipPath := filepath.Join(dir, "bin.zip")
	if err := zipFile(zipPath, "containerkit.exe", src); err != nil {
		t.Fatal(err)
	}
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	if len(zr.File) != 1 || zr.File[0].Name != "containerkit.exe" {
		t.Fatalf("zip files = %+v", zr.File)
	}
	rc, err := zr.File[0].Open()
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	body, err = io.ReadAll(rc)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "containerkit" {
		t.Fatalf("zip body %q", body)
	}
}

func TestCommitFromRepo(t *testing.T) {
	t.Setenv("CONTAINERKIT_COMMIT", "abc123")
	if got := commitFromRepo(t.TempDir()); got != "abc123" {
		t.Fatalf("override: got %q", got)
	}
	t.Setenv("CONTAINERKIT_COMMIT", "")
	if got := commitFromRepo(t.TempDir()); got != "none" {
		t.Fatalf("missing git: got %q", got)
	}
	root, err := moduleRoot()
	if err != nil {
		t.Fatal(err)
	}
	got := commitFromRepo(root)
	if got == "none" || len(got) < 7 {
		t.Fatalf("HEAD: got %q", got)
	}
}

func TestVersionFromRepo(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "VERSION")
	if err := os.WriteFile(path, []byte("1.2.3\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONTAINERKIT_VERSION", "")
	got, err := versionFromRepo(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != "1.2.3" {
		t.Fatalf("VERSION: got %q", got)
	}
	t.Setenv("CONTAINERKIT_VERSION", "9.9.9")
	got, err = versionFromRepo(dir)
	if err != nil || got != "9.9.9" {
		t.Fatalf("override: got %q err %v", got, err)
	}
	if err := os.WriteFile(path, []byte("\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONTAINERKIT_VERSION", "")
	if _, err := versionFromRepo(dir); err == nil {
		t.Fatal("empty VERSION returned a version")
	}
}
