package kit

import (
	"context"
	"testing"
)

func TestLoadProjectNilComposer(t *testing.T) {
	t.Parallel()
	if _, err := LoadProject(context.Background(), nil, Name("app")); err == nil {
		t.Fatal("expected nil composer")
	}
}

func TestLoadProjectNilOption(t *testing.T) {
	t.Parallel()
	if _, err := LoadProject(context.Background(), &fakeComposer{}, Name("app"), nil); err == nil {
		t.Fatal("expected nil option")
	}
}

func TestLoadProjectFluentSetters(t *testing.T) {
	t.Parallel()
	api := &fakeComposer{}
	var heard string
	var unsupported []UnsupportedAttribute
	proj, err := LoadProject(context.Background(), api,
		Name("app"),
		File("compose.yaml"),
		File("compose.override.yaml"),
		ProjectDirectory("/work"),
		EnvFile(".env"),
		Profile("dev"),
		Services("api", "db"),
		AllResources(true),
		Offline(true),
		Compatibility(true),
		InsecureRegistry("localhost:5000"),
		OnLoad(func(event string, _ map[string]any) { heard = event }),
		OnUnsupported(func(attrs []UnsupportedAttribute) { unsupported = attrs }),
		Interpolation(false),
		Normalization(false),
		Consistency(false),
		ResolvedPaths(true),
		DiscardEnvFile(),
		OsEnv(),
		DotEnv(),
		ConfigFileEnv(),
		DefaultConfigPath(),
		EnvironmentResolution(false),
		LabelsResolution(false),
	)
	if err != nil {
		t.Fatal(err)
	}
	if proj == nil || proj.Name != "app" {
		t.Fatalf("project = %+v", proj)
	}
	got := api.loaded
	if got.ProjectName != "app" {
		t.Fatalf("name = %q", got.ProjectName)
	}
	if len(got.ConfigPaths) != 2 || got.ConfigPaths[0] != "compose.yaml" || got.ConfigPaths[1] != "compose.override.yaml" {
		t.Fatalf("files = %#v", got.ConfigPaths)
	}
	if got.WorkingDir != "/work" {
		t.Fatalf("workdir = %q", got.WorkingDir)
	}
	if len(got.EnvFiles) != 1 || got.EnvFiles[0] != ".env" {
		t.Fatalf("env files = %#v", got.EnvFiles)
	}
	if len(got.Profiles) != 1 || got.Profiles[0] != "dev" {
		t.Fatalf("profiles = %#v", got.Profiles)
	}
	if len(got.Services) != 2 || got.Services[0] != "api" {
		t.Fatalf("services = %#v", got.Services)
	}
	if !got.All || !got.Offline || !got.Compatibility {
		t.Fatalf("flags = all=%v offline=%v compat=%v", got.All, got.Offline, got.Compatibility)
	}
	if len(got.OCI.InsecureRegistries) != 1 || got.OCI.InsecureRegistries[0] != "localhost:5000" {
		t.Fatalf("oci = %+v", got.OCI)
	}
	if len(got.LoadListeners) != 1 {
		t.Fatalf("listeners = %d", len(got.LoadListeners))
	}
	got.LoadListeners[0]("extends", nil)
	if heard != "extends" {
		t.Fatalf("listener event = %q", heard)
	}
	if got.OnUnsupportedAttribute == nil {
		t.Fatal("missing OnUnsupportedAttribute")
	}
	got.OnUnsupportedAttribute([]UnsupportedAttribute{{Service: "api", Path: "deploy.endpoint_mode"}})
	if len(unsupported) != 1 || unsupported[0].Service != "api" {
		t.Fatalf("unsupported = %+v", unsupported)
	}
	if len(got.ProjectOptionsFns) < 10 {
		t.Fatalf("project option fns = %d", len(got.ProjectOptionsFns))
	}
}

func TestLoadProjectWorkingDirAndEmptySetters(t *testing.T) {
	t.Parallel()
	api := &fakeComposer{}
	if _, err := LoadProject(context.Background(), api, ProjectDirectory("/src")); err != nil {
		t.Fatal(err)
	}
	if api.loaded.WorkingDir != "/src" {
		t.Fatalf("workdir = %q", api.loaded.WorkingDir)
	}
	tests := []struct {
		name string
		opt  LoadProjectOption
	}{
		{"empty file", File("")},
		{"empty project directory", ProjectDirectory("")},
		{"empty env file", EnvFile("")},
		{"empty profile", Profile("")},
		{"empty registry", InsecureRegistry("")},
		{"nil on load", OnLoad(nil)},
		{"nil on unsupported", OnUnsupported(nil)},
		{"nil project option", ProjectOption(nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if _, err := LoadProject(context.Background(), &fakeComposer{}, tt.opt); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestLoadProjectGrouped(t *testing.T) {
	t.Parallel()
	api := &fakeComposer{}
	if _, err := LoadProject(context.Background(), api, GroupedSet([]any{Name("app"), File("compose.yaml"), Offline(true)})); err != nil {
		t.Fatal(err)
	}
	if api.loaded.ProjectName != "app" || !api.loaded.Offline {
		t.Fatalf("loaded = %+v", api.loaded)
	}
}
