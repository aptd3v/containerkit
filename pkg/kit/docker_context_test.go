package kit

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// TestPackBuildContextDockerignore checks that .dockerignore follows Docker's
// rules: every pattern is anchored at the context root and only ** crosses
// directories (unlike .gitignore, where a slash-free pattern matches at any depth).
func TestPackBuildContextDockerignore(t *testing.T) {
	files := []string{
		"Dockerfile.src",
		"README.md",
		"docs/guide.md",
		"src/main.go",
		"apps/web/index.js",
		"apps/desktop/main.js",
		"apps/desktop/assets/icon.png",
		"a/b/apps/x",
		"ui-tui/src/sdk/apps/index.js",
		"node_modules/left-pad/index.js",
		"ui-tui/node_modules/react/index.js",
		"skills/top.md",
		"skills/tools/web/SKILL.md",
		"skills/tools/web/run.py",
	}

	tests := []struct {
		name   string
		ignore string
		want   []string
	}{
		{
			name:   "slash-free pattern is root anchored",
			ignore: "apps/\n",
			want: []string{
				"Dockerfile.src", "README.md", "a/b/apps/x", "docs/guide.md",
				"node_modules/left-pad/index.js",
				"skills/tools/web/SKILL.md", "skills/tools/web/run.py", "skills/top.md",
				"src/main.go", "ui-tui/node_modules/react/index.js", "ui-tui/src/sdk/apps/index.js",
			},
		},
		{
			name:   "double star matches at any depth",
			ignore: "**/node_modules\n",
			want: []string{
				"Dockerfile.src", "README.md", "a/b/apps/x",
				"apps/desktop/assets/icon.png", "apps/desktop/main.js", "apps/web/index.js",
				"docs/guide.md",
				"skills/tools/web/SKILL.md", "skills/tools/web/run.py", "skills/top.md",
				"src/main.go", "ui-tui/src/sdk/apps/index.js",
			},
		},
		{
			name:   "negation re-includes inside excluded directory",
			ignore: "apps/\n!apps/desktop/**\n",
			want: []string{
				"Dockerfile.src", "README.md", "a/b/apps/x",
				"apps/desktop/assets/icon.png", "apps/desktop/main.js",
				"docs/guide.md", "node_modules/left-pad/index.js",
				"skills/tools/web/SKILL.md", "skills/tools/web/run.py", "skills/top.md",
				"src/main.go", "ui-tui/node_modules/react/index.js", "ui-tui/src/sdk/apps/index.js",
			},
		},
		{
			name:   "root glob does not reach subdirectories",
			ignore: "*.md\n!skills/**/*.md\n",
			want: []string{
				"Dockerfile.src", "a/b/apps/x",
				"apps/desktop/assets/icon.png", "apps/desktop/main.js", "apps/web/index.js",
				"docs/guide.md", "node_modules/left-pad/index.js",
				"skills/tools/web/SKILL.md", "skills/tools/web/run.py", "skills/top.md",
				"src/main.go", "ui-tui/node_modules/react/index.js", "ui-tui/src/sdk/apps/index.js",
			},
		},
		{
			name:   "any-depth glob with negation",
			ignore: "**/*.md\n!skills/**/*.md\n",
			want: []string{
				"Dockerfile.src", "a/b/apps/x",
				"apps/desktop/assets/icon.png", "apps/desktop/main.js", "apps/web/index.js",
				"node_modules/left-pad/index.js",
				"skills/tools/web/SKILL.md", "skills/tools/web/run.py", "skills/top.md",
				"src/main.go", "ui-tui/node_modules/react/index.js", "ui-tui/src/sdk/apps/index.js",
			},
		},
		{
			name:   "leading slash comments and blank lines",
			ignore: "# comment\n\n/src\n  \n/apps/web\n",
			want: []string{
				"Dockerfile.src", "README.md", "a/b/apps/x",
				"apps/desktop/assets/icon.png", "apps/desktop/main.js",
				"docs/guide.md", "node_modules/left-pad/index.js",
				"skills/tools/web/SKILL.md", "skills/tools/web/run.py", "skills/top.md",
				"ui-tui/node_modules/react/index.js", "ui-tui/src/sdk/apps/index.js",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := t.TempDir()
			for _, f := range files {
				p := filepath.Join(src, filepath.FromSlash(f))
				if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(p, []byte(f), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(filepath.Join(src, ".dockerignore"), []byte(tt.ignore), 0o644); err != nil {
				t.Fatal(err)
			}

			r, err := packBuildContext([]byte("FROM scratch\n"), src)
			if err != nil {
				t.Fatal(err)
			}
			got := tarRegularFiles(t, r)
			want := append([]string{".dockerignore", "Dockerfile"}, tt.want...)
			sort.Strings(want)
			if !equalStrings(got, want) {
				t.Fatalf("archive files mismatch\n got: %v\nwant: %v", got, want)
			}
		})
	}
}

func TestPackBuildContextBadDockerignorePattern(t *testing.T) {
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "a.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, ".dockerignore"), []byte("[\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := packBuildContext([]byte("FROM scratch\n"), src); err == nil {
		t.Fatal("expected error for malformed .dockerignore pattern")
	}
}

func tarRegularFiles(t *testing.T, r io.Reader) []string {
	t.Helper()
	tr := tar.NewReader(r)
	var names []string
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if hdr.Typeflag == tar.TypeReg {
			names = append(names, hdr.Name)
		}
	}
	sort.Strings(names)
	return names
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
