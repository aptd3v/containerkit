package catalogtest

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAliasesKeepExactUpstreamNames(t *testing.T) {
	t.Parallel()
	root := moduleRoot(t)
	var mismatches []string
	err := filepath.WalkDir(filepath.Join(root, "internal"), func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Base(path) != "types.go" {
			return err
		}
		if !strings.HasSuffix(filepath.Dir(path), string(filepath.Separator)+"types") && filepath.Base(filepath.Dir(path)) != "types" {
			return nil
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts := spec.(*ast.TypeSpec)
				sel := selectorOf(ts.Type)
				if sel == nil {
					continue
				}
				if ts.Name.Name != sel.Sel.Name {
					rel, _ := filepath.Rel(root, path)
					mismatches = append(mismatches, rel+": "+ts.Name.Name+" aliases "+sel.Sel.Name)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(mismatches) > 0 {
		t.Fatalf("aliases must keep the exact upstream type name:\n  %s", strings.Join(mismatches, "\n  "))
	}
}

func TestCatalogCoversExportedSDKTypes(t *testing.T) {
	t.Parallel()
	aliased := loadAliased(t)
	missing := []string{}
	seen := map[string]bool{}

	for importPath, dir := range sdkPackageDirs(t) {
		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, dir, func(info os.FileInfo) bool {
			name := info.Name()
			return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
		}, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", importPath, err)
		}
		for pkgName, pkg := range pkgs {
			if strings.HasSuffix(pkgName, "_test") {
				continue
			}
			for _, f := range pkg.Files {
				for _, decl := range f.Decls {
					gd, ok := decl.(*ast.GenDecl)
					if !ok || gd.Tok != token.TYPE {
						continue
					}
					for _, spec := range gd.Specs {
						ts := spec.(*ast.TypeSpec)
						if !ts.Name.IsExported() {
							continue
						}
						key := importPath + "." + ts.Name.Name
						if seen[key] {
							continue
						}
						seen[key] = true
						if !aliased[key] {
							missing = append(missing, key)
						}
					}
				}
			}
		}
	}

	if len(missing) > 0 {
		t.Fatalf("catalog missing %d exported SDK types:\n  %s", len(missing), strings.Join(missing, "\n  "))
	}
	if len(seen) < 400 {
		t.Fatalf("expected hundreds of SDK types, found %d", len(seen))
	}
	t.Logf("catalog covers %d exported SDK types under exact names", len(seen))
}

func loadAliased(t *testing.T) map[string]bool {
	t.Helper()
	root := moduleRoot(t)
	out := map[string]bool{}
	err := filepath.WalkDir(filepath.Join(root, "internal"), func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Base(path) != "types.go" {
			return err
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		imports := map[string]string{}
		for _, im := range f.Imports {
			ip := strings.Trim(im.Path.Value, `"`)
			name := defaultImportName(ip)
			if im.Name != nil {
				name = im.Name.Name
			}
			imports[name] = ip
		}
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts := spec.(*ast.TypeSpec)
				sel := selectorOf(ts.Type)
				if sel == nil {
					continue
				}
				id, ok := sel.X.(*ast.Ident)
				if !ok {
					continue
				}
				if ip, ok := imports[id.Name]; ok {
					out[ip+"."+sel.Sel.Name] = true
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) == 0 {
		t.Fatal("no generated aliases found under internal/")
	}
	return out
}

func selectorOf(expr ast.Expr) *ast.SelectorExpr {
	switch e := expr.(type) {
	case *ast.SelectorExpr:
		return e
	case *ast.IndexExpr:
		return selectorOf(e.X)
	case *ast.IndexListExpr:
		return selectorOf(e.X)
	default:
		return nil
	}
}

func sdkPackageDirs(t *testing.T) map[string]string {
	t.Helper()
	sources := []string{
		"github.com/moby/moby/api/types",
		"github.com/moby/moby/api/types/...",
		"github.com/moby/moby/api/pkg/...",
		"github.com/moby/moby/client",
		"github.com/moby/moby/client/pkg/...",
		"github.com/compose-spec/compose-go/v2/...",
		"github.com/docker/compose/v5/pkg/api",
		"github.com/docker/go-connections/...",
		"github.com/docker/go-units",
		"github.com/opencontainers/go-digest",
		"github.com/opencontainers/image-spec/specs-go/v1",
		"github.com/moby/docker-image-spec/specs-go/v1",
		"github.com/distribution/reference",
	}
	args := append([]string{"list", "-e", "-f", "{{.ImportPath}}\t{{.Dir}}"}, sources...)
	cmd := exec.Command("go", args...)
	cmd.Dir = moduleRoot(t)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("go list: %v\n%s", err, stderr.String())
	}
	out := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(stdout.String()), "\n") {
		path, dir, ok := strings.Cut(line, "\t")
		if !ok || dir == "" {
			continue
		}
		if strings.Contains(path, "/internal/") || strings.Contains(path, "/cmd/") {
			continue
		}
		if path == "github.com/compose-spec/compose-go/v2" || path == "github.com/docker/go-connections" {
			continue
		}
		out[path] = dir
	}
	return out
}

func defaultImportName(path string) string {
	base := path
	if i := strings.LastIndex(path, "/"); i >= 0 {
		base = path[i+1:]
	}
	if i := strings.LastIndex(base, "-"); i >= 0 {
		return base[i+1:]
	}
	return base
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for dir := wd; dir != filepath.Dir(dir); dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
	}
	t.Fatal("go.mod not found")
	return ""
}
