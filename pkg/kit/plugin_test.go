package kit

import (
	"archive/tar"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPackPlugin(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		wantErr bool
		want    []string
	}{
		{
			name: "config and rootfs",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"description":"e2e"}`), 0o644); err != nil {
					t.Fatal(err)
				}
				rootfs := filepath.Join(dir, "rootfs")
				if err := os.MkdirAll(filepath.Join(rootfs, "run", "docker", "plugins"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(rootfs, "plugin"), []byte("binary"), 0o755); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			want: []string{"config.json", "rootfs/", "rootfs/plugin", "rootfs/run/", "rootfs/run/docker/", "rootfs/run/docker/plugins/"},
		},
		{
			name: "missing config.json",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				if err := os.Mkdir(filepath.Join(dir, "rootfs"), 0o755); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			wantErr: true,
		},
		{
			name: "missing rootfs",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{}`), 0o644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r, err := PackPlugin(tt.setup(t))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			defer r.Close()
			got := tarNames(t, r)
			for _, name := range tt.want {
				if !got[name] {
					t.Fatalf("missing %s in %v", name, keys(got))
				}
			}
		})
	}
}

func TestPluginOptionsApply(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "enable timeout rounds up subsecond",
			run: func(t *testing.T) {
				t.Helper()
				api := &fakeEngine{}
				if _, err := EnablePlugin(ctx, api, "plug", Timeout(500*time.Millisecond)); err != nil {
					t.Fatal(err)
				}
				if api.pluginEnable.Timeout != 1 {
					t.Fatalf("timeout = %d", api.pluginEnable.Timeout)
				}
			},
		},
		{
			name: "install privileges disabled args",
			run: func(t *testing.T) {
				t.Helper()
				api := &fakeEngine{}
				priv := func(context.Context) (string, error) { return "auth", nil }
				accept := func(context.Context, PluginPrivileges) (bool, error) { return true, nil }
				_, err := InstallPlugin(ctx, api, "plug",
					Name("example.com/plug:latest"),
					Disabled(true),
					AcceptAllPermissions(true),
					Privilege(priv),
					AcceptPermissions(accept),
					Args("DEBUG=1"),
					RegistryAuth("user:pass"),
				)
				if err != nil {
					t.Fatal(err)
				}
				got := api.pluginInstall
				if got.RemoteRef != "example.com/plug:latest" {
					t.Fatalf("remote = %s", got.RemoteRef)
				}
				if !got.Disabled || !got.AcceptAllPermissions {
					t.Fatalf("disabled/accept-all = %+v", got)
				}
				if got.RegistryAuth != "user:pass" {
					t.Fatalf("auth = %s", got.RegistryAuth)
				}
				if got.PrivilegeFunc == nil || got.AcceptPermissionsFunc == nil {
					t.Fatalf("callbacks = %+v", got)
				}
				if len(got.Args) != 1 || got.Args[0] != "DEBUG=1" {
					t.Fatalf("args = %#v", got.Args)
				}
			},
		},
		{
			name: "upgrade privileges disabled args",
			run: func(t *testing.T) {
				t.Helper()
				api := &fakeEngine{}
				priv := func(context.Context) (string, error) { return "", nil }
				accept := func(context.Context, PluginPrivileges) (bool, error) { return false, nil }
				_, err := UpgradePlugin(ctx, api, "plug",
					Name("example.com/plug:v2"),
					Disabled(true),
					AcceptAllPermissions(true),
					Privilege(priv),
					AcceptPermissions(accept),
					Args("DEBUG=0"),
					RegistryAuth("token"),
				)
				if err != nil {
					t.Fatal(err)
				}
				got := api.pluginUpgrade
				if got.RemoteRef != "example.com/plug:v2" || !got.Disabled || !got.AcceptAllPermissions {
					t.Fatalf("upgrade = %+v", got)
				}
				if got.PrivilegeFunc == nil || got.AcceptPermissionsFunc == nil {
					t.Fatalf("callbacks = %+v", got)
				}
				if len(got.Args) != 1 || got.Args[0] != "DEBUG=0" {
					t.Fatalf("args = %#v", got.Args)
				}
			},
		},
		{
			name: "set create list remove disable push",
			run: func(t *testing.T) {
				t.Helper()
				api := &fakeEngine{}
				if _, err := SetPlugin(ctx, api, "plug", Args("DEBUG=1", "FOO=bar")); err != nil {
					t.Fatal(err)
				}
				if len(api.pluginSet.Args) != 2 {
					t.Fatalf("set args = %#v", api.pluginSet.Args)
				}
				body, err := PackPlugin(pluginFixture(t))
				if err != nil {
					t.Fatal(err)
				}
				defer body.Close()
				if _, err := CreatePlugin(ctx, api, Name("containerkit-e2e"), Content(body)); err != nil {
					t.Fatal(err)
				}
				if api.pluginCreate.RepoName != "containerkit-e2e" {
					t.Fatalf("create name = %s", api.pluginCreate.RepoName)
				}
				if _, err := ListPlugins(ctx, api, Filter("enabled", "true")); err != nil {
					t.Fatal(err)
				}
				if api.pluginList.Filters == nil || !api.pluginList.Filters["enabled"]["true"] {
					t.Fatalf("list filters = %+v", api.pluginList.Filters)
				}
				if _, err := RemovePlugin(ctx, api, "plug", Force(true)); err != nil {
					t.Fatal(err)
				}
				if !api.pluginRemove.Force {
					t.Fatal("remove force unset")
				}
				if _, err := DisablePlugin(ctx, api, "plug", Force(true)); err != nil {
					t.Fatal(err)
				}
				if !api.pluginDisable.Force {
					t.Fatal("disable force unset")
				}
				if _, err := PushPlugin(ctx, api, "plug", RegistryAuth("auth")); err != nil {
					t.Fatal(err)
				}
				if api.pluginPush.RegistryAuth != "auth" {
					t.Fatalf("push auth = %s", api.pluginPush.RegistryAuth)
				}
			},
		},
		{
			name: "create wraps engine error",
			run: func(t *testing.T) {
				t.Helper()
				if _, err := CreatePlugin(ctx, nil, Name("plug")); !errors.Is(err, errRequired) {
					t.Fatalf("err = %v", err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.run(t)
		})
	}
}

func pluginFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"description":"fixture"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "rootfs"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func tarNames(t *testing.T, r io.Reader) map[string]bool {
	t.Helper()
	tr := tar.NewReader(r)
	out := map[string]bool{}
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return out
		}
		if err != nil {
			t.Fatal(err)
		}
		out[hdr.Name] = true
	}
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
