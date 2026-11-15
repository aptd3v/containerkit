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

// ImportHit is one import of path from an internal package.
type ImportHit struct {
	Path    string
	FromPkg string
	File    string
	Types   []string
}

// PkgCount is how often an import path appears, which internal packages use it,
// and which exported types in those packages alias it.
type PkgCount struct {
	Path  string
	Count int
	From  []string
	Types []string
}

// Group is similar import paths that share a Go module (or fallback prefix).
type Group struct {
	Name  string
	Count int
	Pkgs  []PkgCount
}

func collectImports(root, module, internalRel string) ([]ImportHit, error) {
	internalDir := filepath.Join(root, internalRel)
	info, err := os.Stat(internalDir)
	if err != nil {
		return nil, fmt.Errorf("internal: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", internalDir)
	}

	var hits []ImportHit
	err = filepath.WalkDir(internalDir, func(path string, d os.DirEntry, err error) error {
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
		relPkg, err := filepath.Rel(root, filepath.Dir(path))
		if err != nil {
			return err
		}
		fromPkg := module + "/" + filepath.ToSlash(relPkg)
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		aliasToPath, dotImports := importAliases(f)
		typesByImport := exportedTypesByImport(f, aliasToPath, dotImports)
		seen := map[string]bool{}
		for _, im := range f.Imports {
			ip := strings.Trim(im.Path.Value, `"`)
			if seen[ip] {
				continue
			}
			seen[ip] = true
			hits = append(hits, ImportHit{
				Path:    ip,
				FromPkg: fromPkg,
				File:    path,
				Types:   typesByImport[ip],
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return hits, nil
}

func parseModules(root string) ([]string, error) {
	b, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return nil, fmt.Errorf("read go.mod: %w", err)
	}
	var modules []string
	inRequire := false
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if i := strings.Index(line, "//"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" {
			continue
		}
		switch {
		case strings.HasPrefix(line, "module "):
			modules = append(modules, strings.TrimSpace(strings.TrimPrefix(line, "module ")))
		case line == "require (":
			inRequire = true
		case inRequire && line == ")":
			inRequire = false
		case strings.HasPrefix(line, "require "):
			if p := moduleFromRequire(strings.TrimPrefix(line, "require ")); p != "" {
				modules = append(modules, p)
			}
		case inRequire:
			if p := moduleFromRequire(line); p != "" {
				modules = append(modules, p)
			}
		}
	}
	sort.Slice(modules, func(i, j int) bool {
		if len(modules[i]) != len(modules[j]) {
			return len(modules[i]) > len(modules[j])
		}
		return modules[i] < modules[j]
	})
	return modules, nil
}

func moduleFromRequire(line string) string {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return ""
	}
	p := fields[0]
	if !strings.Contains(p, ".") && !strings.Contains(p, "/") {
		return ""
	}
	return p
}

func similarGroup(importPath string, modules []string) string {
	for _, m := range modules {
		if importPath == m || strings.HasPrefix(importPath, m+"/") {
			return m
		}
	}
	return fallbackGroup(importPath)
}

func fallbackGroup(importPath string) string {
	first, _, _ := strings.Cut(importPath, "/")
	if !strings.Contains(first, ".") {
		return "stdlib"
	}
	parts := strings.Split(importPath, "/")
	n := 3
	if len(parts) < n {
		return importPath
	}
	return strings.Join(parts[:n], "/")
}

func importAliases(f *ast.File) (map[string]string, []string) {
	aliasToPath := map[string]string{}
	var dotImports []string
	for _, im := range f.Imports {
		ip := strings.Trim(im.Path.Value, `"`)
		if im.Name != nil && im.Name.Name == "." {
			dotImports = append(dotImports, ip)
			continue
		}
		if im.Name != nil && im.Name.Name == "_" {
			continue
		}
		aliasToPath[importName(im, ip)] = ip
	}
	return aliasToPath, dotImports
}

func importName(im *ast.ImportSpec, ip string) string {
	if im.Name != nil && im.Name.Name != "." && im.Name.Name != "_" {
		return im.Name.Name
	}
	return defaultImportName(ip)
}

func defaultImportName(ip string) string {
	base := ip
	if i := strings.LastIndex(ip, "/"); i >= 0 {
		base = ip[i+1:]
	}
	if i := strings.LastIndex(base, "-"); i >= 0 {
		return base[i+1:]
	}
	return base
}

func exportedTypesByImport(f *ast.File, aliasToPath map[string]string, dotImports []string) map[string][]string {
	out := map[string][]string{}
	seen := map[string]map[string]bool{}
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
			ip := importOfExpr(ts.Type, aliasToPath, dotImports)
			if ip == "" {
				continue
			}
			if seen[ip] == nil {
				seen[ip] = map[string]bool{}
			}
			if seen[ip][ts.Name.Name] {
				continue
			}
			seen[ip][ts.Name.Name] = true
			out[ip] = append(out[ip], ts.Name.Name)
		}
	}
	for ip := range out {
		sort.Strings(out[ip])
	}
	return out
}

func importOfExpr(expr ast.Expr, aliasToPath map[string]string, dotImports []string) string {
	switch e := expr.(type) {
	case *ast.Ident:
		if len(dotImports) == 1 {
			return dotImports[0]
		}
		return ""
	case *ast.SelectorExpr:
		id, ok := e.X.(*ast.Ident)
		if !ok {
			return ""
		}
		return aliasToPath[id.Name]
	case *ast.StarExpr:
		return importOfExpr(e.X, aliasToPath, dotImports)
	case *ast.ParenExpr:
		return importOfExpr(e.X, aliasToPath, dotImports)
	case *ast.IndexExpr:
		return importOfExpr(e.X, aliasToPath, dotImports)
	case *ast.IndexListExpr:
		return importOfExpr(e.X, aliasToPath, dotImports)
	default:
		return ""
	}
}

func groupImports(hits []ImportHit, modules []string) []Group {
	type agg struct {
		count int
		from  map[string]bool
		types map[string]bool
	}
	byPath := map[string]*agg{}
	for _, h := range hits {
		a := byPath[h.Path]
		if a == nil {
			a = &agg{from: map[string]bool{}, types: map[string]bool{}}
			byPath[h.Path] = a
		}
		a.count++
		a.from[h.FromPkg] = true
		for _, name := range h.Types {
			a.types[name] = true
		}
	}

	byGroup := map[string][]PkgCount{}
	groupTotal := map[string]int{}
	for path, a := range byPath {
		from := make([]string, 0, len(a.from))
		for p := range a.from {
			from = append(from, p)
		}
		sort.Strings(from)
		types := make([]string, 0, len(a.types))
		for name := range a.types {
			types = append(types, name)
		}
		sort.Strings(types)
		g := similarGroup(path, modules)
		byGroup[g] = append(byGroup[g], PkgCount{Path: path, Count: a.count, From: from, Types: types})
		groupTotal[g] += a.count
	}

	groups := make([]Group, 0, len(byGroup))
	for name, pkgs := range byGroup {
		sort.Slice(pkgs, func(i, j int) bool {
			if pkgs[i].Count != pkgs[j].Count {
				return pkgs[i].Count > pkgs[j].Count
			}
			return pkgs[i].Path < pkgs[j].Path
		})
		groups = append(groups, Group{Name: name, Count: groupTotal[name], Pkgs: pkgs})
	}
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].Count != groups[j].Count {
			return groups[i].Count > groups[j].Count
		}
		return groups[i].Name < groups[j].Name
	})
	return groups
}

func filterGroups(groups []Group, query string) []Group {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return groups
	}
	var out []Group
	for _, g := range groups {
		groupMatch := strings.Contains(strings.ToLower(g.Name), q)
		var pkgs []PkgCount
		count := 0
		for _, p := range g.Pkgs {
			pathMatch := groupMatch || strings.Contains(strings.ToLower(p.Path), q)
			fromMatch := false
			for _, from := range p.From {
				if strings.Contains(strings.ToLower(from), q) {
					fromMatch = true
					break
				}
			}
			var types []string
			if pathMatch || fromMatch {
				types = p.Types
			} else {
				for _, name := range p.Types {
					if strings.Contains(strings.ToLower(name), q) {
						types = append(types, name)
					}
				}
			}
			if pathMatch || fromMatch || len(types) > 0 {
				p.Types = types
				pkgs = append(pkgs, p)
				count += p.Count
			}
		}
		if len(pkgs) == 0 {
			continue
		}
		g.Pkgs = pkgs
		g.Count = count
		out = append(out, g)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func formatReport(groups []Group, total int, query string) string {
	var b strings.Builder
	if q := strings.TrimSpace(query); q != "" {
		fmt.Fprintf(&b, "pkg imports in internal/: %d unique paths in %d similar groups matching %q (most used first)\n", total, len(groups), q)
	} else {
		fmt.Fprintf(&b, "pkg imports in internal/: %d unique paths in %d similar groups (most used first)\n", total, len(groups))
	}
	if len(groups) == 0 {
		return b.String()
	}
	for _, g := range groups {
		fmt.Fprintf(&b, "\n%s  (%d)\n", g.Name, g.Count)
		for _, p := range g.Pkgs {
			fmt.Fprintf(&b, "  %4d  %s\n", p.Count, p.Path)
			for _, from := range p.From {
				fmt.Fprintf(&b, "        %s\n", from)
			}
			for _, name := range p.Types {
				fmt.Fprintf(&b, "          %s\n", name)
			}
		}
	}
	return b.String()
}

func uniquePathCount(groups []Group) int {
	n := 0
	for _, g := range groups {
		n += len(g.Pkgs)
	}
	return n
}
