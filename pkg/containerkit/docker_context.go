package containerkit

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// NewLocalBuildContext returns a tar build context for client.ImageBuild.
// src must be a directory. The archive contains the generated Dockerfile,
// skips a root Dockerfile already present in src, honors .dockerignore,
// and preserves file modes and symlinks.
func (d *dockerFile) NewLocalBuildContext(src string) (io.Reader, error) {
	if err := d.Validate(); err != nil {
		return nil, err
	}
	info, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("source directory %s does not exist", src)
		}
		return nil, fmt.Errorf("source directory %s: %w", src, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("source %s is not a directory", src)
	}
	rules, err := readDockerignore(src)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	finished := false
	defer func() {
		if !finished {
			_ = tw.Close()
		}
	}()

	body := []byte(d.String())
	if err := writeTarBytes(tw, "Dockerfile", 0644, body); err != nil {
		return nil, err
	}

	err = filepath.Walk(src, func(filePath string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if filePath == src {
			return nil
		}
		rel, err := filepath.Rel(src, filePath)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "Dockerfile" {
			return nil
		}
		excluded, err := rules.excluded(rel, info.IsDir())
		if err != nil {
			return err
		}
		if excluded {
			return nil
		}
		return addContextEntry(tw, filePath, rel, info)
	})
	if err != nil {
		return nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	finished = true
	return &buf, nil
}

func addContextEntry(tw *tar.Writer, filePath, rel string, info os.FileInfo) error {
	mode := info.Mode()
	switch {
	case mode&os.ModeSymlink != 0:
		target, err := os.Readlink(filePath)
		if err != nil {
			return fmt.Errorf("read symlink %s: %w", rel, err)
		}
		return tw.WriteHeader(&tar.Header{
			Name:     rel,
			Mode:     int64(mode.Perm()),
			Typeflag: tar.TypeSymlink,
			Linkname: target,
			ModTime:  info.ModTime(),
		})
	case info.IsDir():
		return tw.WriteHeader(&tar.Header{
			Name:     rel + "/",
			Mode:     int64(mode.Perm()),
			Typeflag: tar.TypeDir,
			ModTime:  info.ModTime(),
		})
	case mode.IsRegular():
		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		return writeTarBytes(tw, rel, mode.Perm(), data)
	default:
		return fmt.Errorf("unsupported file %s", rel)
	}
}

func writeTarBytes(tw *tar.Writer, name string, mode os.FileMode, data []byte) error {
	header := &tar.Header{
		Name:     name,
		Size:     int64(len(data)),
		Mode:     int64(mode),
		Typeflag: tar.TypeReg,
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	_, err := tw.Write(data)
	return err
}

type ignorePattern struct {
	pattern  string
	negate   bool
	dirOnly  bool
	anchored bool
}

type ignoreRules struct {
	patterns []ignorePattern
}

func readDockerignore(dir string) (ignoreRules, error) {
	raw, err := os.ReadFile(filepath.Join(dir, ".dockerignore"))
	if err != nil {
		if os.IsNotExist(err) {
			return ignoreRules{}, nil
		}
		return ignoreRules{}, fmt.Errorf("read .dockerignore: %w", err)
	}
	var rules ignoreRules
	for _, line := range strings.Split(string(raw), "\n") {
		pattern, ok := parseIgnoreLine(line)
		if !ok {
			continue
		}
		rules.patterns = append(rules.patterns, pattern)
	}
	return rules, nil
}

func parseIgnoreLine(line string) (ignorePattern, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return ignorePattern{}, false
	}
	p := ignorePattern{}
	if strings.HasPrefix(line, "!") {
		p.negate = true
		line = strings.TrimSpace(strings.TrimPrefix(line, "!"))
		if line == "" {
			return ignorePattern{}, false
		}
	}
	if strings.HasSuffix(line, "/") {
		p.dirOnly = true
		line = strings.TrimSuffix(line, "/")
	}
	line = strings.TrimPrefix(line, "/")
	if line == "" {
		return ignorePattern{}, false
	}
	p.pattern = line
	p.anchored = strings.Contains(line, "/")
	return p, true
}

func (r ignoreRules) excluded(rel string, isDir bool) (bool, error) {
	rel = strings.TrimPrefix(filepath.ToSlash(rel), "/")
	excluded := false
	for _, pattern := range r.patterns {
		matched, err := pattern.matches(rel, isDir)
		if err != nil {
			return false, err
		}
		if matched {
			excluded = !pattern.negate
		}
	}
	return excluded, nil
}

func (p ignorePattern) matches(rel string, isDir bool) (bool, error) {
	candidates := []string{rel}
	if !p.anchored {
		parts := strings.Split(rel, "/")
		for i := 1; i < len(parts); i++ {
			candidates = append(candidates, strings.Join(parts[i:], "/"))
		}
	}
	for _, candidate := range candidates {
		if !p.dirOnly {
			returnMatch, err := globMatch(p.pattern, candidate)
			if err != nil || returnMatch {
				return returnMatch, err
			}
			continue
		}
		segs := strings.Split(candidate, "/")
		for i := 1; i <= len(segs); i++ {
			prefix := strings.Join(segs[:i], "/")
			ok, err := globMatch(p.pattern, prefix)
			if err != nil {
				return false, err
			}
			if !ok {
				continue
			}
			if i == len(segs) && !isDir {
				continue
			}
			return true, nil
		}
	}
	return false, nil
}

func globMatch(pattern, name string) (bool, error) {
	return matchSeg(strings.Split(pattern, "/"), strings.Split(name, "/"), pattern)
}

func matchSeg(pat, name []string, original string) (bool, error) {
	for {
		if len(pat) == 0 {
			return len(name) == 0, nil
		}
		if pat[0] == "**" {
			ok, err := matchSeg(pat[1:], name, original)
			if err != nil || ok {
				return ok, err
			}
			if len(name) == 0 {
				return false, nil
			}
			name = name[1:]
			continue
		}
		if len(name) == 0 {
			return false, nil
		}
		ok, err := path.Match(pat[0], name[0])
		if err != nil {
			return false, fmt.Errorf("dockerignore pattern %q: %w", original, err)
		}
		if !ok {
			return false, nil
		}
		pat = pat[1:]
		name = name[1:]
	}
}
