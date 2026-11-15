// Command unusedtypes lists exported types under internal/ that pkg/kit
// does not reference. Use it while wrapping the catalog to see what is left.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "unusedtypes: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	fail := flag.Bool("fail", false, "exit 1 when unused types remain")
	kitRel := flag.String("kit", "pkg/kit", "directory whose references count as used")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: unusedtypes [flags]\n\n")
		fmt.Fprintf(os.Stderr, "List exported types under internal/ that are not referenced from pkg/kit.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	root, err := findRoot()
	if err != nil {
		return err
	}
	mod, err := modulePath(root)
	if err != nil {
		return err
	}
	catalog, err := catalogTypes(root, mod)
	if err != nil {
		return err
	}
	used, err := referencedTypes(filepath.Join(root, *kitRel), mod)
	if err != nil {
		return err
	}
	unused := unusedTypes(catalog, used)
	fmt.Fprint(os.Stdout, formatReport(unused, catalog))
	if *fail && len(unused) > 0 {
		return fmt.Errorf("%d unused types", len(unused))
	}
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
