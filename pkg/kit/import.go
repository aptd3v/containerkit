package kit

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"strings"

	loadertypes "github.com/aptd3v/containerkit/internal/compose/loader/types"
	composetypes "github.com/aptd3v/containerkit/internal/compose/types"
)

type importAct struct {
	name     string
	workDir  string
	files    []string
	env      map[string]string
	profiles []string
}

// ImportYAML loads a Compose file from YAML bytes.
// Parse errors accumulate on the returned [Compose]; call [Compose.Validate].
func ImportYAML(yml []byte, opts ...ImportOption) *Compose {
	act, err := collectImport(opts)
	if err != nil {
		return importErr(err)
	}
	if len(yml) == 0 {
		return importErr(cfgErr("compose", "import", "ImportYAML", fmt.Errorf("%w: yaml", errEmpty)))
	}
	name := act.name
	if name == "" {
		name = "compose.yaml"
	}
	details := composetypes.ConfigDetails{
		WorkingDir: importWorkDir(act),
		ConfigFiles: []composetypes.ConfigFile{{
			Filename: name,
			Content:  append([]byte(nil), yml...),
		}},
		Environment: importEnv(act),
	}
	details.ConfigFiles = append(details.ConfigFiles, importPathFiles(act.files)...)
	return loadCompose(details, act)
}

// ImportFile loads a Compose file from path. Additional [File] options are
// overlay compose files (compose.yaml + compose.override.yaml).
func ImportFile(path string, opts ...ImportOption) *Compose {
	act, err := collectImport(opts)
	if err != nil {
		return importErr(err)
	}
	if path == "" {
		return importErr(cfgErr("compose", "import", "ImportFile", fmt.Errorf("%w: path", errEmpty)))
	}
	if act.workDir == "" {
		act.workDir = filepath.Dir(path)
	}
	files := append([]string{path}, act.files...)
	details := composetypes.ConfigDetails{
		WorkingDir:  importWorkDir(act),
		ConfigFiles: importPathFiles(files),
		Environment: importEnv(act),
	}
	return loadCompose(details, act)
}

// Import loads Compose YAML from r.
func Import(r io.Reader, opts ...ImportOption) *Compose {
	if r == nil {
		return importErr(cfgErr("compose", "import", "Import", fmt.Errorf("%w: reader", errRequired)))
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return importErr(cfgErr("compose", "import", "Import", err))
	}
	return ImportYAML(buf.Bytes(), opts...)
}

func collectImport(opts []ImportOption) (importAct, error) {
	var act importAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = append(errs, cfgErr("compose", fmt.Sprintf("import.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyImport(&act); err != nil {
			errs = append(errs, wrapPath("import", err))
		}
	}
	return act, join(errs...)
}

func loadCompose(details composetypes.ConfigDetails, act importAct) *Compose {
	proj, err := loadertypes.LoadWithContext(context.Background(), details, func(o *loadertypes.Options) {
		if act.name != "" {
			o.SetProjectName(act.name, true)
		}
		if len(act.profiles) > 0 {
			o.Profiles = append(o.Profiles, act.profiles...)
		}
	})
	if err != nil {
		resource := act.name
		if resource == "" {
			resource = details.WorkingDir
		}
		if resource == "" {
			resource = "import"
		}
		return importErr(cfgErr("compose", resource, "Import", err))
	}
	// The loader already kept the services for these profiles. Leaving the
	// profile tags on them makes a later Up hide the same services again,
	// and depends_on then fails with "no such service".
	if len(act.profiles) > 0 {
		for name, svc := range proj.Services {
			svc.Profiles = nil
			proj.Services[name] = svc
		}
	}
	p, err := composeFromProject(proj)
	if err != nil {
		return importErr(err)
	}
	return p
}

func importErr(err error) *Compose {
	p := &Compose{}
	p.errs = addErr(p.errs, wrapReceiver("compose", err))
	return p
}

func importWorkDir(act importAct) string {
	if act.workDir != "" {
		return act.workDir
	}
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

func importEnv(act importAct) composetypes.Mapping {
	env := composetypes.Mapping{}
	for _, kv := range os.Environ() {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		env[k] = v
	}
	maps.Copy(env, act.env)
	return env
}

func importPathFiles(paths []string) []composetypes.ConfigFile {
	out := make([]composetypes.ConfigFile, 0, len(paths))
	for _, p := range paths {
		if p == "" {
			continue
		}
		out = append(out, composetypes.ConfigFile{Filename: p})
	}
	return out
}

func (o NameOpt) applyImport(a *importAct) error {
	if err := validateName(o.name); err != nil {
		return err
	}
	a.name = o.name
	return nil
}

func (o ProjectDirectoryOpt) applyImport(a *importAct) error {
	if o.s == "" {
		return cfgErr("", "project_directory", "ProjectDirectory", fmt.Errorf("%w: project directory", errEmpty))
	}
	a.workDir = o.s
	return nil
}

func (o FileOpt) applyImport(a *importAct) error {
	if o.path == "" {
		return cfgErr("", "file", "File", fmt.Errorf("%w: compose file", errEmpty))
	}
	a.files = append(a.files, o.path)
	return nil
}

func (o EnvOpt) applyImport(a *importAct) error {
	if o.key == "" {
		return cfgErr("", "env", "Env", fmt.Errorf("%w: env key", errEmpty))
	}
	if a.env == nil {
		a.env = map[string]string{}
	}
	a.env[o.key] = o.value
	return nil
}

func (o ProfileOpt) applyImport(a *importAct) error {
	if o.p == "" {
		return cfgErr("", "profile", "Profile", fmt.Errorf("%w: profile", errEmpty))
	}
	a.profiles = append(a.profiles, o.p)
	return nil
}
