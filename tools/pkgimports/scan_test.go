package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGroupImportsSortsMostUsedSimilarFirst(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		files      map[string]string
		wantGroups []string
		wantCounts []int
		wantFirst  []string
	}{
		{
			name: "empty internal has no groups",
			files: map[string]string{
				"internal/.keep": "",
			},
			wantGroups: nil,
		},
		{
			name: "similar modules grouped and ranked by usage",
			files: map[string]string{
				"internal/engine/types/types.go":      "package types\nimport \"github.com/moby/moby/api/types\"\n",
				"internal/engine/image/types.go":      "package types\nimport \"github.com/moby/moby/api/types/image\"\n",
				"internal/engine/container/t.go":      "package types\nimport \"github.com/moby/moby/api/types/container\"\n",
				"internal/compose/types/types.go":     "package types\nimport \"github.com/compose-spec/compose-go/v2/types\"\n",
				"internal/units/types/types.go":       "package types\nimport \"github.com/docker/go-units\"\n",
				"internal/catalogtest/x.go":           "package catalogtest\nimport \"testing\"\n",
				"internal/engine/types/types_test.go": "package types\nimport \"testing\"\n",
			},
			wantGroups: []string{
				"github.com/moby/moby/api",
				"github.com/compose-spec/compose-go/v2",
				"github.com/docker/go-units",
			},
			wantCounts: []int{3, 1, 1},
			wantFirst: []string{
				"github.com/moby/moby/api/types",
				"github.com/moby/moby/api/types/container",
				"github.com/moby/moby/api/types/image",
			},
		},
		{
			name: "same import from two packages counts twice",
			files: map[string]string{
				"internal/a/a.go": "package a\nimport \"github.com/docker/go-units\"\n",
				"internal/b/b.go": "package b\nimport \"github.com/docker/go-units\"\n",
				"internal/c/c.go": "package c\nimport \"github.com/moby/moby/api/types\"\n",
			},
			wantGroups: []string{
				"github.com/docker/go-units",
				"github.com/moby/moby/api",
			},
			wantCounts: []int{2, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := writeImportFixture(t, tt.files)
			mod, err := modulePath(root)
			if err != nil {
				t.Fatal(err)
			}
			modules, err := parseModules(root)
			if err != nil {
				t.Fatal(err)
			}
			hits, err := collectImports(root, mod, "internal")
			if err != nil {
				t.Fatal(err)
			}
			groups := groupImports(hits, modules)
			if len(groups) != len(tt.wantGroups) {
				t.Fatalf("groups = %v, want %v", groupNames(groups), tt.wantGroups)
			}
			for i, name := range tt.wantGroups {
				if groups[i].Name != name {
					t.Fatalf("group[%d] = %s, want %s (order %v)", i, groups[i].Name, name, groupNames(groups))
				}
				if tt.wantCounts != nil && groups[i].Count != tt.wantCounts[i] {
					t.Fatalf("group %s count = %d, want %d", name, groups[i].Count, tt.wantCounts[i])
				}
			}
			if len(tt.wantFirst) > 0 {
				got := make([]string, 0, len(groups[0].Pkgs))
				for _, p := range groups[0].Pkgs {
					got = append(got, p.Path)
				}
				if strings.Join(got, ",") != strings.Join(tt.wantFirst, ",") {
					t.Fatalf("first group pkgs = %v, want %v", got, tt.wantFirst)
				}
			}
		})
	}
}

func TestFormatReportMostUsedFirst(t *testing.T) {
	t.Parallel()
	groups := []Group{
		{
			Name:  "github.com/moby/moby/api",
			Count: 2,
			Pkgs: []PkgCount{
				{Path: "github.com/moby/moby/api/types", Count: 1, From: []string{"example.com/mod/internal/engine/types"}},
				{Path: "github.com/moby/moby/api/types/image", Count: 1, From: []string{"example.com/mod/internal/engine/image"}},
			},
		},
		{
			Name:  "github.com/docker/go-units",
			Count: 1,
			Pkgs: []PkgCount{
				{Path: "github.com/docker/go-units", Count: 1, From: []string{"example.com/mod/internal/units/types"}},
			},
		},
	}
	got := formatReport(groups, uniquePathCount(groups), "")
	apiAt := strings.Index(got, "github.com/moby/moby/api  (2)")
	unitsAt := strings.Index(got, "github.com/docker/go-units  (1)")
	if apiAt < 0 || unitsAt < 0 || apiAt > unitsAt {
		t.Fatalf("expected api group before go-units:\n%s", got)
	}
	if !strings.Contains(got, "pkg imports in internal/: 3 unique paths in 2 similar groups") {
		t.Fatalf("summary missing: %s", got)
	}
}

func TestCollectsAndPrintsTypeNames(t *testing.T) {
	t.Parallel()
	root := writeImportFixture(t, map[string]string{
		"internal/connections/nat/types/types.go": `package types
import "github.com/docker/go-connections/nat"
type Port = nat.Port
type PortMap = nat.PortMap
`,
		"internal/units/types/types.go": `package types
import "github.com/docker/go-units"
type Ulimit = units.Ulimit
`,
	})
	mod, err := modulePath(root)
	if err != nil {
		t.Fatal(err)
	}
	modules, err := parseModules(root)
	if err != nil {
		t.Fatal(err)
	}
	hits, err := collectImports(root, mod, "internal")
	if err != nil {
		t.Fatal(err)
	}
	groups := groupImports(hits, modules)
	got := formatReport(groups, uniquePathCount(groups), "")
	for _, name := range []string{"Port", "PortMap", "Ulimit"} {
		if !strings.Contains(got, "          "+name+"\n") {
			t.Fatalf("type %s missing from report:\n%s", name, got)
		}
	}
}

func TestSearchFilter(t *testing.T) {
	t.Parallel()

	files := map[string]string{
		"internal/connections/nat/types/types.go": `package types
import "github.com/docker/go-connections/nat"
type Port = nat.Port
type PortMap = nat.PortMap
`,
		"internal/engine/image/types/types.go": `package types
import "github.com/moby/moby/api/types/image"
type Summary = image.Summary
`,
		"internal/units/types/types.go": `package types
import "github.com/docker/go-units"
type Ulimit = units.Ulimit
`,
	}

	tests := []struct {
		name        string
		query       string
		wantGroups  []string
		wantTypes   []string
		hideTypes   []string
		wantSummary string
	}{
		{
			name:       "type name Port",
			query:      "Port",
			wantGroups: []string{"github.com/docker/go-connections"},
			wantTypes:  []string{"Port", "PortMap"},
			hideTypes:  []string{"Summary", "Ulimit"},
		},
		{
			name:       "case-insensitive type",
			query:      "ulimit",
			wantGroups: []string{"github.com/docker/go-units"},
			wantTypes:  []string{"Ulimit"},
			hideTypes:  []string{"Port", "Summary"},
		},
		{
			name:       "import path fragment",
			query:      "moby/api",
			wantGroups: []string{"github.com/moby/moby/api"},
			wantTypes:  []string{"Summary"},
			hideTypes:  []string{"Port", "Ulimit"},
		},
		{
			name:        "no matches",
			query:       "does-not-exist",
			wantGroups:  nil,
			wantSummary: `matching "does-not-exist"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := writeImportFixture(t, files)
			mod, err := modulePath(root)
			if err != nil {
				t.Fatal(err)
			}
			modules, err := parseModules(root)
			if err != nil {
				t.Fatal(err)
			}
			hits, err := collectImports(root, mod, "internal")
			if err != nil {
				t.Fatal(err)
			}
			groups := filterGroups(groupImports(hits, modules), tt.query)
			if got := groupNames(groups); strings.Join(got, ",") != strings.Join(tt.wantGroups, ",") {
				t.Fatalf("groups = %v, want %v", got, tt.wantGroups)
			}
			got := formatReport(groups, uniquePathCount(groups), tt.query)
			if tt.wantSummary != "" && !strings.Contains(got, tt.wantSummary) {
				t.Fatalf("summary missing %q:\n%s", tt.wantSummary, got)
			}
			for _, name := range tt.wantTypes {
				if !strings.Contains(got, "          "+name+"\n") {
					t.Fatalf("expected type %s:\n%s", name, got)
				}
			}
			for _, name := range tt.hideTypes {
				if strings.Contains(got, "          "+name+"\n") {
					t.Fatalf("unexpected type %s:\n%s", name, got)
				}
			}
		})
	}
}

func TestRepoInternalImportRanking(t *testing.T) {
	root, err := findRoot()
	if err != nil {
		t.Fatal(err)
	}
	mod, err := modulePath(root)
	if err != nil {
		t.Fatal(err)
	}
	modules, err := parseModules(root)
	if err != nil {
		t.Fatal(err)
	}
	hits, err := collectImports(root, mod, "internal")
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) < 40 {
		t.Fatalf("expected catalog imports, got %d", len(hits))
	}
	groups := groupImports(hits, modules)
	if len(groups) < 2 {
		t.Fatal("expected multiple similar import groups")
	}
	for i := 1; i < len(groups); i++ {
		if groups[i-1].Count < groups[i].Count {
			t.Fatalf("groups not sorted most-used first: %s (%d) then %s (%d)",
				groups[i-1].Name, groups[i-1].Count, groups[i].Name, groups[i].Count)
		}
	}
	if groups[0].Name != "github.com/moby/moby/api" {
		t.Fatalf("most used similar imports = %s (%d), want github.com/moby/moby/api", groups[0].Name, groups[0].Count)
	}
	var sawPort bool
	for _, g := range groups {
		for _, p := range g.Pkgs {
			for _, name := range p.Types {
				if name == "Port" && strings.Contains(p.Path, "go-connections/nat") {
					sawPort = true
				}
			}
		}
	}
	if !sawPort {
		t.Fatal("expected type Port on github.com/docker/go-connections/nat")
	}
	filtered := filterGroups(groups, "Port")
	if len(filtered) == 0 {
		t.Fatal("search Port returned no groups")
	}
	for _, g := range filtered {
		for _, p := range g.Pkgs {
			for _, name := range p.Types {
				if !strings.Contains(strings.ToLower(name), "port") && !strings.Contains(strings.ToLower(p.Path), "port") && !strings.Contains(strings.ToLower(g.Name), "port") {
					t.Fatalf("search Port kept unrelated type %s in %s", name, p.Path)
				}
			}
		}
	}
}

func writeImportFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	gomod := `module example.com/mod

go 1.22

require (
	github.com/compose-spec/compose-go/v2 v2.0.0
	github.com/docker/go-units v0.5.0
	github.com/moby/moby/api v1.0.0
)
`
	mustWriteFile(t, filepath.Join(root, "go.mod"), gomod)
	for rel, body := range files {
		mustWriteFile(t, filepath.Join(root, rel), body)
	}
	return root
}

func groupNames(groups []Group) []string {
	out := make([]string, 0, len(groups))
	for _, g := range groups {
		out = append(out, g.Name)
	}
	return out
}

func mustWriteFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
