package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnusedTypesFromTestdata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		used   []string
		unused []string
	}{
		{
			name:   "no kit references",
			used:   nil,
			unused: []string{"Alpha", "Beta"},
		},
		{
			name:   "one type referenced",
			used:   []string{"Alpha"},
			unused: []string{"Beta"},
		},
		{
			name:   "all types referenced",
			used:   []string{"Alpha", "Beta"},
			unused: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := writeFixture(t, tt.used)
			mod, err := modulePath(root)
			if err != nil {
				t.Fatal(err)
			}
			catalog, err := catalogTypes(root, mod)
			if err != nil {
				t.Fatal(err)
			}
			if len(catalog) != 2 {
				t.Fatalf("catalogued %d types, want 2", len(catalog))
			}
			for _, c := range catalog {
				if c.Name != "Alpha" && c.Name != "Beta" {
					t.Fatalf("catalog must keep exact names, got %s", c.Name)
				}
			}
			used, err := referencedTypes(filepath.Join(root, "pkg", "kit"), mod)
			if err != nil {
				t.Fatal(err)
			}
			unused := unusedTypes(catalog, used)
			got := names(unused)
			if !equalNames(got, tt.unused) {
				t.Fatalf("unused = %v, want %v (used keys %v)", got, tt.unused, keys(used))
			}
		})
	}
}

func TestRepoCatalogKeepsExactNamesAndScansKit(t *testing.T) {
	root, err := findRoot()
	if err != nil {
		t.Fatal(err)
	}
	mod, err := modulePath(root)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := catalogTypes(root, mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog) < 400 {
		t.Fatalf("expected hundreds of catalogued types, got %d", len(catalog))
	}
	used, err := referencedTypes(filepath.Join(root, "pkg", "kit"), mod)
	if err != nil {
		t.Fatal(err)
	}
	unused := unusedTypes(catalog, used)
	if len(used) == 0 && len(unused) != len(catalog) {
		t.Fatalf("with no pkg/kit references, unused=%d catalog=%d", len(unused), len(catalog))
	}
	var sawComposeConfig bool
	for _, c := range catalog {
		if strings.HasSuffix(c.ImportPath, "/internal/compose/types") && c.Name == "Config" {
			sawComposeConfig = true
		}
		if c.Name == "ComposeConfig" {
			t.Fatalf("catalog must keep exact names, found %s", c.Name)
		}
	}
	if !sawComposeConfig {
		t.Fatal("expected exact name Config in internal/compose/types")
	}
}

func TestFormatReportListsExactNames(t *testing.T) {
	t.Parallel()
	catalog := []TypeRef{
		{ImportPath: "example.com/mod/internal/foo/types", Name: "Alpha"},
		{ImportPath: "example.com/mod/internal/foo/types", Name: "Beta"},
	}
	unused := []TypeRef{catalog[1]}
	got := formatReport(unused, catalog)
	if !strings.Contains(got, "unused types: 1 / 2") {
		t.Fatalf("summary missing: %s", got)
	}
	if !strings.Contains(got, "  Beta\n") {
		t.Fatalf("exact unused name missing: %s", got)
	}
	if strings.Contains(got, "Alpha") {
		t.Fatalf("used type listed as unused: %s", got)
	}
}

func writeFixture(t *testing.T, used []string) string {
	t.Helper()
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "go.mod"), "module example.com/mod\n\ngo 1.22\n")
	mustWrite(t, filepath.Join(root, "internal", "foo", "types", "types.go"), `package types

type Alpha struct{}
type Beta struct{}
`)
	body := "package kit\n"
	if len(used) > 0 {
		body += `
import footypes "example.com/mod/internal/foo/types"
`
		for _, name := range used {
			body += "type Use" + name + " = footypes." + name + "\n"
		}
	}
	mustWrite(t, filepath.Join(root, "pkg", "kit", "use.go"), body)
	return root
}

func names(refs []TypeRef) []string {
	out := make([]string, 0, len(refs))
	for _, r := range refs {
		out = append(out, r.Name)
	}
	return out
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func equalNames(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func mustWrite(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
