package kit

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/moby/patternmatcher"
	"github.com/moby/patternmatcher/ignorefile"
)

// Archive returns a tar build context for [BuildImage].
// contextDir is packed when non-empty; an empty directory packs only this Dockerfile.
// A .dockerignore in contextDir is honored. The fluent Dockerfile is written as "Dockerfile".
func (d *Dockerfile) Archive(contextDir string) (io.Reader, error) {
	if err := d.Validate(); err != nil {
		return nil, wrapPath("dockerfile", err)
	}
	return packBuildContext([]byte(d.String()), contextDir)
}

func packBuildContext(dockerfile []byte, src string) (io.Reader, error) {
	if src == "" {
		var buf bytes.Buffer
		tw := tar.NewWriter(&buf)
		if err := writeTarBytes(tw, "Dockerfile", 0644, dockerfile); err != nil {
			_ = tw.Close()
			return nil, err
		}
		if err := tw.Close(); err != nil {
			return nil, err
		}
		return &buf, nil
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
	ignore, err := readDockerignore(src)
	if err != nil {
		return nil, err
	}
	// Match results per visited directory, so children reuse their parent's
	// results the same way the Docker CLI walks a context.
	dirMatches := map[string]patternmatcher.MatchInfo{}

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	finished := false
	defer func() {
		if !finished {
			_ = tw.Close()
		}
	}()

	if err := writeTarBytes(tw, "Dockerfile", 0644, dockerfile); err != nil {
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
		if ignore != nil {
			excluded, match, err := ignore.MatchesUsingParentResults(rel, dirMatches[path.Dir(rel)])
			if err != nil {
				return fmt.Errorf(".dockerignore: %w", err)
			}
			if info.IsDir() {
				dirMatches[rel] = match
			}
			if excluded {
				// Without "!" patterns nothing below an excluded directory can
				// come back, so skip it. Otherwise keep walking for re-includes.
				if info.IsDir() && !ignore.Exclusions() {
					return filepath.SkipDir
				}
				return nil
			}
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

// readDockerignore loads contextDir/.dockerignore with Docker's own parser and
// matcher, so patterns are anchored at the context root and only ** spans
// directories. It returns nil when there is no .dockerignore.
func readDockerignore(dir string) (*patternmatcher.PatternMatcher, error) {
	f, err := os.Open(filepath.Join(dir, ".dockerignore"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read .dockerignore: %w", err)
	}
	defer f.Close()
	patterns, err := ignorefile.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read .dockerignore: %w", err)
	}
	if len(patterns) == 0 {
		return nil, nil
	}
	pm, err := patternmatcher.New(patterns)
	if err != nil {
		return nil, fmt.Errorf(".dockerignore: %w", err)
	}
	return pm, nil
}
