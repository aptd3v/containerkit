package codegen

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
)

func TestGenerate_minimalProject(t *testing.T) {
	project := &types.Project{
		Name: "test-project",
		Services: types.Services{
			"web": types.ServiceConfig{
				Name:  "web",
				Image: "nginx:alpine",
				Ports: []types.ServicePortConfig{
					{
						Protocol:  "tcp",
						HostIP:    "0.0.0.0",
						Published: "8080",
						Target:    80,
					},
				},
			},
		},
		Networks: types.Networks{},
		Volumes:  types.Volumes{},
	}

	out, err := Generate(project, Options{PackageName: "main", EmitMain: false})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	s := string(out)

	// Must contain key go-contain API usage
	for _, substr := range []string{
		"create.NewProject",
		"WithService",
		"create.NewContainer",
		"cc.WithImage",
		"hc.WithPortBindings",
		"nginx:alpine",
		"8080",
		"80",
		"test-project",
	} {
		if !strings.Contains(s, substr) {
			t.Errorf("generated code missing %q", substr)
		}
	}
}

func TestGenerate_withMain(t *testing.T) {
	project := &types.Project{
		Name: "runme",
		Services: types.Services{
			"svc": types.ServiceConfig{
				Name:  "svc",
				Image: "alpine:latest",
			},
		},
		Networks: types.Networks{},
		Volumes:  types.Volumes{},
	}

	out, err := Generate(project, Options{PackageName: "main", EmitMain: true})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	s := string(out)

	for _, substr := range []string{
		"func main()",
		"compose.NewCompose",
		"app.Up",
		"up.WithWriter",
		"up.WithRemoveOrphans",
		"signal.Notify",
		"app.Down",
		"down.WithRemoveOrphans",
		"context.Canceled",
	} {
		if !strings.Contains(s, substr) {
			t.Errorf("generated code with EmitMain missing %q", substr)
		}
	}
}

// TestGenerate_e2eRealComposeFile loads a real docker-compose file from testdata,
// generates Go code, and verifies the output compiles.
func TestGenerate_e2eRealComposeFile(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(filename)
	composePath := filepath.Join(testDir, "testdata", "docker-compose.yaml")
	if _, err := os.Stat(composePath); err != nil {
		t.Fatalf("testdata compose file missing: %v", err)
	}

	// Set env vars so ${VAR} in docker-compose.yaml resolve (same as testdata/.env for manual runs).
	restore := setEnv(map[string]string{
		"POSTGRES_IMAGE":   "postgres:16-alpine",
		"POSTGRES_USER":   "pguser",
		"POSTGRES_PASSWORD": "secure_password_here",
		"POSTGRES_DB":     "app",
		"PGPORT":          "5432",
		"NGINX_TAG":       "alpine",
		"APP_PORT":        "3000",
		"REDIS_IMAGE":     "redis:7-alpine",
	})
	defer restore()

	opts, err := cli.NewProjectOptions(
		[]string{composePath},
		cli.WithOsEnv,
		cli.WithDotEnv,
		cli.WithWorkingDirectory(filepath.Dir(composePath)),
		cli.WithProfiles([]string{"full"}), // include worker (profile: full) so generated code has WithDependsOnHealthy
	)
	if err != nil {
		t.Fatalf("NewProjectOptions: %v", err)
	}

	ctx := context.Background()
	project, err := opts.LoadProject(ctx)
	if err != nil {
		t.Fatalf("LoadProject: %v", err)
	}

	out, err := Generate(project, Options{PackageName: "main", EmitMain: true})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	s := string(out)

	// Assert generated code reflects the real compose file (testdata/docker-compose.yaml)
	for _, substr := range []string{
		"some-stack",
		"api",
		"db",
		"redis",
		"curler",
		"worker",
		"backend",
		"frontend",
		"postgres_data",
		"redis_data",
		"api-cache",
		"postgres:16-alpine",
		"nginx",
		"curlimages/curl",
		"cc.WithImage",
		"hc.WithPortBindings",
		"sc.WithDependsOn",
		"sc.WithDependsOnHealthy",
		"cc.WithHealthCheck",
		"hc.WithRWNamedVolumeMount",
		"network.WithDriver",
		"deploy.WithReplicas",
	} {
		if !strings.Contains(s, substr) {
			t.Errorf("e2e generated code missing %q", substr)
		}
	}

	// Optionally verify generated code compiles (requires temp module and replace).
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.go")
	if err := os.WriteFile(mainPath, out, 0644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
	moduleRoot := filepath.Clean(filepath.Join(testDir, "..", ".."))
	replaceDir, err := filepath.Abs(moduleRoot)
	if err != nil {
		t.Fatalf("abs module root: %v", err)
	}
	goMod := "module e2etest\n\ngo 1.23\n\nrequire github.com/aptd3v/go-contain v0.0.0\n\nreplace github.com/aptd3v/go-contain => " + replaceDir + "\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = dir
	if tidyOut, err := tidy.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy: %v\n%s", err, tidyOut)
	}
	cmd := exec.Command("go", "build", "-o", filepath.Join(dir, "e2e"), ".")
	cmd.Dir = dir
	if buildOut, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed (generated code must compile): %v\n%s", err, buildOut)
	}
}

const envUnsetSentinel = "\x00<unset>"

// setEnv sets the given env vars and returns a func that restores the previous state.
func setEnv(m map[string]string) (restore func()) {
	prev := make(map[string]string)
	for k, v := range m {
		if old, ok := os.LookupEnv(k); ok {
			prev[k] = old
		} else {
			prev[k] = envUnsetSentinel
		}
		os.Setenv(k, v)
	}
	return func() {
		for k, old := range prev {
			if old == envUnsetSentinel {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, old)
			}
		}
	}
}
