package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aptd3v/go-contain/internal/codegen"
	"github.com/compose-spec/compose-go/v2/cli"
)

type stringSlice []string

func (s *stringSlice) String() string { return fmt.Sprintf("%v", *s) }
func (s *stringSlice) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func main() {
	var (
		output      = flag.String("o", "", "output Go file path (default: stdout)")
		pkg         = flag.String("pkg", "main", "package name for generated code")
		emitMain    = flag.Bool("main", false, "emit func main() that runs compose.Up and defers Down")
		projectName = flag.String("project", "", "override project name in generated code")
		envPath     = flag.String("env", "", "path to .env file; if set, use only this env file for all -f files")
		help        = flag.Bool("help", false, "show usage and exit")
		profiles    stringSlice
		configFiles stringSlice
	)
	flag.Var(&configFiles, "f", "compose file path (can be repeated)")
	flag.Var(&profiles, "profile", "compose profile to include (can be repeated); only services with these profiles are generated")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: go-contain-codegen [flags] [compose files...]\n")
		fmt.Fprintf(os.Stderr, "  Reads docker-compose YAML and generates go-contain Go code.\n\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n  If no compose files are given, defaults to docker-compose.yaml and compose.yaml.\n")
	}
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	configPaths := configFiles
	if len(configPaths) == 0 {
		configPaths = flag.Args()
	}
	if len(configPaths) == 0 {
		flag.Usage()
		os.Exit(0)
	}

	// Load env: either the single file from -env, or .env from each -f file's directory.
	if *envPath != "" {
		path, err := filepath.Abs(*envPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "go-contain-codegen: -env: %v\n", err)
			os.Exit(1)
		}
		loadEnvFile(path)
	} else {
		for _, p := range configPaths {
			if composeDir, err := filepath.Abs(filepath.Dir(p)); err == nil {
				loadEnvFile(filepath.Join(composeDir, ".env"))
			}
		}
	}
	optsFuncs := []cli.ProjectOptionsFn{cli.WithOsEnv, cli.WithDotEnv}
	if len(configPaths) > 0 {
		if wd, err := filepath.Abs(filepath.Dir(configPaths[0])); err == nil {
			optsFuncs = append(optsFuncs, cli.WithWorkingDirectory(wd))
		}
	}
	if len(profiles) > 0 {
		optsFuncs = append(optsFuncs, cli.WithProfiles([]string(profiles)))
	}
	opts, err := cli.NewProjectOptions(configPaths, optsFuncs...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "go-contain-codegen: %v\n", err)
		os.Exit(1)
	}
	if *projectName != "" {
		opts.Name = *projectName
	}

	ctx := context.Background()
	project, err := opts.LoadProject(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "go-contain-codegen: load project: %v\n", err)
		os.Exit(1)
	}
	if len(profiles) > 0 {
		project, err = project.WithProfiles(profiles)
		if err != nil {
			fmt.Fprintf(os.Stderr, "go-contain-codegen: filter by profiles: %v\n", err)
			os.Exit(1)
		}
	}

	out, err := codegen.Generate(project, codegen.Options{
		PackageName: *pkg,
		EmitMain:    *emitMain,
		ProjectName: *projectName,
		Profiles:    []string(profiles),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "go-contain-codegen: generate: %v\n", err)
		os.Exit(1)
	}

	if *output == "" {
		_, _ = os.Stdout.Write(out)
		return
	}
	if err := os.WriteFile(*output, out, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "go-contain-codegen: write %s: %v\n", *output, err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Wrote %s\n", *output)
}

// loadEnvFile reads a .env file and sets KEY=VALUE into the process environment.
// Skips empty lines and lines starting with #. Missing file is ignored.
func loadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		i := strings.Index(line, "=")
		if i <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:i])
		value := strings.TrimSpace(line[i+1:])
		if key == "" {
			continue
		}
		// Remove surrounding quotes if present
		if len(value) >= 2 && (value[0] == '"' && value[len(value)-1] == '"' || value[0] == '\'' && value[len(value)-1] == '\'') {
			value = value[1 : len(value)-1]
		}
		if _, ok := os.LookupEnv(key); ok {
			fmt.Fprintf(os.Stderr, "The %q variable is being overwritten by %s\n", key, path)
		}
		_ = os.Setenv(key, value)
	}
}
