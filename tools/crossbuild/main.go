// Command crossbuild compiles cmd/containerkit for linux, darwin, and windows
// on amd64 and arm64 into dist/ as archives plus checksums.txt, and copies each
// binary into npm/@aptd3v/containerkit-<os>-<arch>/bin. It does not publish.
package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var targets = []struct{ goos, goarch string }{
	{"linux", "amd64"},
	{"linux", "arm64"},
	{"darwin", "amd64"},
	{"darwin", "arm64"},
	{"windows", "amd64"},
	{"windows", "arm64"},
}

func main() {
	root, err := moduleRoot()
	if err != nil {
		fatal(err)
	}
	dist := filepath.Join(root, "dist")
	if err := os.RemoveAll(dist); err != nil {
		fatal(err)
	}
	if err := os.MkdirAll(dist, 0o755); err != nil {
		fatal(err)
	}
	tmp, err := os.MkdirTemp("", "containerkit-crossbuild-")
	if err != nil {
		fatal(err)
	}
	defer os.RemoveAll(tmp)

	version, err := versionFromRepo(root)
	if err != nil {
		fatal(err)
	}
	commit := commitFromRepo(root)
	ld := fmt.Sprintf("-s -w -X main.version=%s -X main.commit=%s", version, commit)
	pkg := "./cmd/containerkit"

	var archives []string
	for _, t := range targets {
		binName := "containerkit"
		if t.goos == "windows" {
			binName += ".exe"
		}
		binPath := filepath.Join(tmp, t.goos+"_"+t.goarch, binName)
		if err := os.MkdirAll(filepath.Dir(binPath), 0o755); err != nil {
			fatal(err)
		}
		cmd := exec.Command("go", "build", "-trimpath", "-ldflags", ld, "-o", binPath, pkg)
		cmd.Dir = root
		cmd.Env = append(os.Environ(),
			"CGO_ENABLED=0",
			"GOOS="+t.goos,
			"GOARCH="+t.goarch,
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		fmt.Fprintf(os.Stderr, "build %s/%s\n", t.goos, t.goarch)
		if err := cmd.Run(); err != nil {
			fatal(err)
		}
		notice := filepath.Join(root, "NOTICE")
		entries := []archiveEntry{
			{name: binName, src: binPath, mode: 0o755},
			{name: "NOTICE", src: notice, mode: 0o644},
		}
		var archive string
		if t.goos == "windows" {
			archive = filepath.Join(dist, fmt.Sprintf("containerkit_%s_%s.zip", t.goos, t.goarch))
			if err := zipFile(archive, entries); err != nil {
				fatal(err)
			}
		} else {
			archive = filepath.Join(dist, fmt.Sprintf("containerkit_%s_%s.tar.gz", t.goos, t.goarch))
			if err := tarball(archive, entries); err != nil {
				fatal(err)
			}
		}
		archives = append(archives, archive)
		npmBin := filepath.Join(root, "npm", "@aptd3v", "containerkit-"+t.goos+"-"+t.goarch, "bin", binName)
		if err := os.MkdirAll(filepath.Dir(npmBin), 0o755); err != nil {
			fatal(err)
		}
		if err := copyFile(binPath, npmBin); err != nil {
			fatal(err)
		}
		if err := os.Chmod(npmBin, 0o755); err != nil {
			fatal(err)
		}
		npmNotice := filepath.Join(root, "npm", "@aptd3v", "containerkit-"+t.goos+"-"+t.goarch, "NOTICE")
		if err := copyFile(notice, npmNotice); err != nil {
			fatal(err)
		}
		if t.goos == runtime.GOOS && t.goarch == runtime.GOARCH {
			native := filepath.Join(dist, binName)
			if err := copyFile(binPath, native); err != nil {
				fatal(err)
			}
			if err := os.Chmod(native, 0o755); err != nil {
				fatal(err)
			}
		}
	}
	if err := writeChecksums(filepath.Join(dist, "checksums.txt"), archives); err != nil {
		fatal(err)
	}
	fmt.Fprintf(os.Stderr, "wrote %s (%d archives, go %s)\n", dist, len(archives), runtime.Version())
}

type archiveEntry struct {
	name string
	src  string
	mode int64
}

func tarball(archive string, entries []archiveEntry) (err error) {
	f, err := os.Create(archive)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}()
	gz := gzip.NewWriter(f)
	defer func() {
		if cerr := gz.Close(); err == nil {
			err = cerr
		}
	}()
	tw := tar.NewWriter(gz)
	defer func() {
		if cerr := tw.Close(); err == nil {
			err = cerr
		}
	}()
	for _, entry := range entries {
		if err := writeTar(tw, entry); err != nil {
			return err
		}
	}
	return nil
}

func writeTar(tw *tar.Writer, entry archiveEntry) error {
	in, err := os.Open(entry.src)
	if err != nil {
		return err
	}
	defer in.Close()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	mode := entry.mode
	if mode == 0 {
		mode = 0o644
	}
	if err := tw.WriteHeader(&tar.Header{
		Name: entry.name,
		Mode: mode,
		Size: info.Size(),
	}); err != nil {
		return err
	}
	_, err = io.Copy(tw, in)
	return err
}

func zipFile(archive string, entries []archiveEntry) (err error) {
	f, err := os.Create(archive)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}()
	zw := zip.NewWriter(f)
	defer func() {
		if cerr := zw.Close(); err == nil {
			err = cerr
		}
	}()
	for _, entry := range entries {
		if err := writeZip(zw, entry); err != nil {
			return err
		}
	}
	return nil
}

func writeZip(zw *zip.Writer, entry archiveEntry) error {
	in, err := os.Open(entry.src)
	if err != nil {
		return err
	}
	defer in.Close()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	hdr, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	hdr.Name = entry.name
	hdr.Method = zip.Deflate
	if entry.mode != 0 {
		hdr.SetMode(os.FileMode(entry.mode))
	}
	w, err := zw.CreateHeader(hdr)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, in)
	return err
}

func writeChecksums(path string, files []string) error {
	var b strings.Builder
	for _, file := range files {
		sum, err := sha256File(file)
		if err != nil {
			return err
		}
		fmt.Fprintf(&b, "%s  %s\n", sum, filepath.Base(file))
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func moduleRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

func commitFromRepo(root string) string {
	if v := strings.TrimSpace(os.Getenv("CONTAINERKIT_COMMIT")); v != "" {
		return v
	}
	cmd := exec.Command("git", "-C", root, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "none"
	}
	commit := strings.TrimSpace(string(out))
	if commit == "" {
		return "none"
	}
	return commit
}

func versionFromRepo(root string) (string, error) {
	if v := strings.TrimSpace(os.Getenv("CONTAINERKIT_VERSION")); v != "" {
		return v, nil
	}
	b, err := os.ReadFile(filepath.Join(root, "VERSION"))
	if err != nil {
		return "", fmt.Errorf("read VERSION: %w", err)
	}
	v := strings.TrimSpace(string(b))
	if v == "" || strings.ContainsAny(v, " \t\r\n") {
		return "", fmt.Errorf("VERSION must be a single version")
	}
	return v, nil
}

func fatal(v any) {
	fmt.Fprintln(os.Stderr, v)
	os.Exit(1)
}
