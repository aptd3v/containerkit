package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// TypeRef is an exported catalog type under internal/<domain>/types.
type TypeRef struct {
	ImportPath string
	Name       string
	File       string
}

func (t TypeRef) Key() string {
	return t.ImportPath + "." + t.Name
}

func modulePath(root string) (string, error) {
	b, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("read go.mod: %w", err)
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("no module path in go.mod")
}

func catalogTypes(root, module string) ([]TypeRef, error) {
	internalDir := filepath.Join(root, "internal")
	var out []TypeRef
	err := filepath.WalkDir(internalDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "catalogtest" || d.Name() == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		if filepath.Base(filepath.Dir(path)) != "types" {
			return nil
		}
		rel, err := filepath.Rel(root, filepath.Dir(path))
		if err != nil {
			return err
		}
		importPath := module + "/" + filepath.ToSlash(rel)
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
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
				out = append(out, TypeRef{
					ImportPath: importPath,
					Name:       ts.Name.Name,
					File:       path,
				})
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ImportPath != out[j].ImportPath {
			return out[i].ImportPath < out[j].ImportPath
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

func referencedTypes(kitDir, module string) (map[string]bool, error) {
	used := map[string]bool{}
	info, err := os.Stat(kitDir)
	if err != nil {
		return nil, fmt.Errorf("pkg/kit: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", kitDir)
	}

	fset := token.NewFileSet()
	err = filepath.WalkDir(kitDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		aliasToPath := map[string]string{}
		var dotImports []string
		internalPrefix := module + "/internal/"
		for _, im := range f.Imports {
			ip := strings.Trim(im.Path.Value, `"`)
			if ip != module+"/internal" && !strings.HasPrefix(ip, internalPrefix) {
				continue
			}
			name := importName(im, ip)
			if im.Name != nil && im.Name.Name == "." {
				dotImports = append(dotImports, ip)
				continue
			}
			if im.Name != nil && im.Name.Name == "_" {
				continue
			}
			aliasToPath[name] = ip
		}
		ast.Inspect(f, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			id, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}
			if ip, ok := aliasToPath[id.Name]; ok {
				used[ip+"."+sel.Sel.Name] = true
			}
			return true
		})
		if len(dotImports) > 0 {
			ast.Inspect(f, func(n ast.Node) bool {
				id, ok := n.(*ast.Ident)
				if !ok || id.Name == "" || !ast.IsExported(id.Name) {
					return true
				}
				for _, ip := range dotImports {
					used[ip+"."+id.Name] = true
				}
				return true
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return used, nil
}

func importName(im *ast.ImportSpec, ip string) string {
	if im.Name != nil && im.Name.Name != "." && im.Name.Name != "_" {
		return im.Name.Name
	}
	base := ip
	if i := strings.LastIndex(ip, "/"); i >= 0 {
		base = ip[i+1:]
	}
	return base
}

func unusedTypes(catalog []TypeRef, used map[string]bool) []TypeRef {
	var out []TypeRef
	for _, t := range catalog {
		if !used[t.Key()] {
			out = append(out, t)
		}
	}
	return out
}

func formatReport(unused, catalog []TypeRef) string {
	var b strings.Builder
	fmt.Fprintf(&b, "unused types: %d / %d catalogued in internal/ are not referenced from pkg/kit\n", len(unused), len(catalog))
	if len(unused) == 0 {
		return b.String()
	}
	current := ""
	for _, t := range unused {
		if t.ImportPath != current {
			current = t.ImportPath
			fmt.Fprintf(&b, "\n%s\n", current)
		}
		fmt.Fprintf(&b, "  %s\n", t.Name)
	}
	return b.String()
}
