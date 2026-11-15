//go:build e2e

package plugin_test

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/aptd3v/containerkit/pkg/kit"
)

const (
	pluginName = "containerkit/e2e-volume:latest"
	volumeName = "containerkit-e2e-vol"
)

func TestEnginePluginLifecycle(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	t.Cleanup(cancel)

	api, err := kit.NewEngine()
	if err != nil {
		t.Skipf("NewEngine: %v", err)
	}
	t.Cleanup(func() { _ = api.Close() })
	if _, err := kit.Ping(ctx, api); err != nil {
		t.Skipf("engine not reachable: %v", err)
	}

	info, err := kit.Info(ctx, api)
	if err != nil {
		t.Fatalf("Info: %v", err)
	}
	if strings.EqualFold(info.Info.OSType, "windows") {
		t.Skip("managed plugins are not supported on Windows Engine")
	}

	dir := assemblePlugin(t, goarch(info.Info.Architecture))
	body, err := kit.PackPlugin(dir)
	if err != nil {
		t.Fatalf("PackPlugin: %v", err)
	}
	t.Cleanup(func() { _ = body.Close() })

	cleanupPlugin(ctx, t, api)
	t.Cleanup(func() { cleanupPlugin(ctx, t, api) })

	if _, err := kit.CreatePlugin(ctx, api, kit.Name(pluginName), kit.Content(body)); err != nil {
		t.Fatalf("CreatePlugin: %v", err)
	}

	if _, err := kit.SetPlugin(ctx, api, pluginName, kit.Args("DEBUG=1")); err != nil {
		t.Fatalf("SetPlugin: %v", err)
	}

	before, err := kit.InspectPlugin(ctx, api, pluginName)
	if err != nil {
		t.Fatalf("InspectPlugin before enable: %v", err)
	}
	if before.Plugin.Enabled {
		t.Fatal("plugin enabled before EnablePlugin")
	}
	if !containsEnv(before.Plugin.Settings.Env, "DEBUG=1") {
		t.Fatalf("env after SetPlugin = %#v", before.Plugin.Settings.Env)
	}

	if _, err := kit.EnablePlugin(ctx, api, pluginName, kit.Timeout(60*time.Second)); err != nil {
		t.Fatalf("EnablePlugin: %v", err)
	}

	after, err := kit.InspectPlugin(ctx, api, pluginName)
	if err != nil {
		t.Fatalf("InspectPlugin after enable: %v", err)
	}
	if !after.Plugin.Enabled {
		t.Fatal("plugin not enabled")
	}

	listed, err := kit.ListPlugins(ctx, api, kit.Filter("enabled", "true"))
	if err != nil {
		t.Fatalf("ListPlugins: %v", err)
	}
	if !pluginListed(listed.Items, pluginName) {
		t.Fatalf("ListPlugins missing %s", pluginName)
	}

	vol, err := kit.CreateVolume(ctx, api,
		kit.Name(volumeName),
		kit.Driver(pluginName),
	)
	if err != nil {
		t.Fatalf("CreateVolume: %v", err)
	}
	t.Cleanup(func() {
		_, _ = kit.RemoveVolume(context.Background(), api, volumeName, kit.Force(true))
	})
	if vol.Volume.Name != volumeName {
		t.Fatalf("volume name = %s", vol.Volume.Name)
	}
	if vol.Volume.Driver != pluginName && vol.Volume.Driver != strings.TrimSuffix(pluginName, ":latest") {
		t.Fatalf("volume driver = %s", vol.Volume.Driver)
	}
	inspected, err := kit.InspectVolume(ctx, api, volumeName)
	if err != nil {
		t.Fatalf("InspectVolume: %v", err)
	}
	if inspected.Volume.Driver != pluginName {
		t.Fatalf("volume driver = %s", inspected.Volume.Driver)
	}
	if _, err := kit.RemoveVolume(ctx, api, volumeName, kit.Force(true)); err != nil {
		t.Fatalf("RemoveVolume: %v", err)
	}

	t.Run("install invalid remote", func(t *testing.T) {
		_, err := kit.InstallPlugin(ctx, api, "containerkit-e2e-missing",
			kit.Name("not a valid remote"),
			kit.Disabled(true),
			kit.AcceptAllPermissions(true),
			kit.Privilege(func(context.Context) (string, error) { return "", nil }),
			kit.AcceptPermissions(func(context.Context, kit.PluginPrivileges) (bool, error) { return true, nil }),
			kit.Args("DEBUG=1"),
		)
		if err == nil {
			t.Fatal("expected InstallPlugin error")
		}
	})
	t.Run("upgrade invalid remote", func(t *testing.T) {
		rc, err := kit.UpgradePlugin(ctx, api, pluginName,
			kit.Name("not a valid remote"),
			kit.Disabled(true),
			kit.AcceptAllPermissions(true),
			kit.Privilege(func(context.Context) (string, error) { return "", nil }),
			kit.AcceptPermissions(func(context.Context, kit.PluginPrivileges) (bool, error) { return true, nil }),
			kit.Args("DEBUG=0"),
		)
		if rc != nil {
			_, _ = io.Copy(io.Discard, rc)
			_ = rc.Close()
		}
		if err == nil {
			t.Fatal("expected UpgradePlugin error")
		}
	})

	if _, err := kit.DisablePlugin(ctx, api, pluginName, kit.Force(true)); err != nil {
		t.Fatalf("DisablePlugin: %v", err)
	}
	disabled, err := kit.InspectPlugin(ctx, api, pluginName)
	if err != nil {
		t.Fatalf("InspectPlugin after disable: %v", err)
	}
	if disabled.Plugin.Enabled {
		t.Fatal("plugin still enabled")
	}
	if _, err := kit.RemovePlugin(ctx, api, pluginName, kit.Force(true)); err != nil {
		t.Fatalf("RemovePlugin: %v", err)
	}
}

func assemblePlugin(t *testing.T, arch string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Dir(file)
	dir := t.TempDir()
	rootfs := filepath.Join(dir, "rootfs")
	for _, p := range []string{
		filepath.Join(rootfs, "run", "docker", "plugins"),
		filepath.Join(rootfs, "data"),
	} {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	cfg, err := os.ReadFile(filepath.Join(root, "testdata", "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), cfg, 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "build", "-o", filepath.Join(rootfs, "plugin"), filepath.Join(root, "volumeplugin"))
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS=linux", "GOARCH="+arch)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build plugin: %v\n%s", err, out)
	}
	return dir
}

func goarch(engineArch string) string {
	switch strings.ToLower(engineArch) {
	case "aarch64", "arm64":
		return "arm64"
	case "x86_64", "amd64":
		return "amd64"
	default:
		return runtime.GOARCH
	}
}

func cleanupPlugin(ctx context.Context, t *testing.T, api kit.Engine) {
	t.Helper()
	_, _ = kit.RemoveVolume(ctx, api, volumeName, kit.Force(true))
	_, _ = kit.DisablePlugin(ctx, api, pluginName, kit.Force(true))
	_, _ = kit.RemovePlugin(ctx, api, pluginName, kit.Force(true))
}

func pluginListed(items []kit.Plugin, name string) bool {
	for _, p := range items {
		if p.Name == name || p.Name+":latest" == name || name == p.Name+":latest" {
			return true
		}
	}
	return false
}

func containsEnv(env []string, want string) bool {
	for _, e := range env {
		if e == want {
			return true
		}
	}
	return false
}
