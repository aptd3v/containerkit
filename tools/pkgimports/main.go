// Command pkgimports lists package imports under internal/, grouped by
// similar Go module, sorted from most used to least used.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "pkgimports: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	internalRel := flag.String("internal", "internal", "directory to scan for imports")
	search := flag.String("search", "", "filter groups, import paths, importers, and type names (case-insensitive)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: pkgimports [flags] [query]\n\n")
		fmt.Fprintf(os.Stderr, "Print package imports under internal/, with type names, grouped by similar module, most used first.\n")
		fmt.Fprintf(os.Stderr, "A search query matches module groups, import paths, importer packages, and type names.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	query := strings.TrimSpace(*search)
	if query == "" && flag.NArg() > 0 {
		query = strings.Join(flag.Args(), " ")
	}

	root, err := findRoot()
	if err != nil {
		return err
	}
	mod, err := modulePath(root)
	if err != nil {
		return err
	}
	modules, err := parseModules(root)
	if err != nil {
		return err
	}
	hits, err := collectImports(root, mod, *internalRel)
	if err != nil {
		return err
	}
	groups := filterGroups(groupImports(hits, modules), query)
	fmt.Fprint(os.Stdout, formatReport(groups, uniquePathCount(groups), query))
	return nil
}

func findRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for dir := wd; dir != filepath.Dir(dir); dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
	}
	return "", fmt.Errorf("go.mod not found from %s", wd)
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
