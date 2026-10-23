//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aptd3v/containerkit/pkg/client"
)

const sshfsPlugin = "vieux/sshfs"

func TestClientPlugin(t *testing.T) {
	t.Parallel()
	cli := newCLI(t)
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()

	if _, err := cli.PluginList(ctx, nil); err != nil {
		t.Fatalf("PluginList nil: %v", err)
	}
	if _, err := cli.PluginList(ctx, &client.PluginList{}); err != nil {
		t.Fatalf("PluginList: %v", err)
	}
	if _, err := cli.PluginInspect(ctx, "containerkit-e2e-missing"); !client.IsErrNotFound(err) {
		t.Fatalf("PluginInspect missing: want not found, got %v", err)
	}

	alias := ident(t, "pl")
	rc, err := cli.PluginInstall(ctx, alias, &client.PluginInstall{
		RemoteRef:            sshfsPlugin,
		AcceptAllPermissions: true,
	})
	if err != nil {
		t.Skipf("PluginInstall %s as %s: %v", sshfsPlugin, alias, err)
	}
	if _, copyErr := io.Copy(testWriter(t), rc); copyErr != nil {
		fmt.Fprintln(testWriter(t), "PluginInstall stream:", copyErr)
	}
	rc.Close()
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = cli.PluginDisable(c, alias, &client.PluginDisable{Force: true})
		_ = cli.PluginRemove(c, alias, &client.PluginRemove{Force: true})
	})

	insp, err := cli.PluginInspect(ctx, alias)
	if err != nil {
		t.Skipf("PluginInspect after install %s: %v", alias, err)
	}
	if insp == nil || insp.Name == "" {
		t.Fatal("PluginInspect: empty plugin")
	}
	fmt.Fprintf(testWriter(t), "plugin name=%s enabled=%v ref=%s\n", insp.Name, insp.Enabled, insp.PluginReference)
	if len(insp.Raw) == 0 {
		t.Fatal("PluginInspect: empty raw")
	}

	if !insp.Enabled {
		if err := cli.PluginEnable(ctx, alias, &client.PluginEnable{Timeout: 30}); err != nil {
			t.Fatalf("PluginEnable: %v", err)
		}
	}
	listed, err := cli.PluginList(ctx, nil)
	if err != nil {
		t.Fatalf("PluginList after install: %v", err)
	}
	found := false
	for _, p := range listed {
		if p.Name == alias || p.Name == insp.Name || strings.Contains(p.Name, alias) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("PluginList: %s missing", alias)
	}

	if err := cli.PluginDisable(ctx, alias, &client.PluginDisable{Force: true}); err != nil {
		t.Fatalf("PluginDisable: %v", err)
	}
	if err := cli.PluginSet(ctx, alias, &client.PluginSet{Args: []string{"DEBUG=1"}}); err != nil {
		fmt.Fprintln(testWriter(t), "PluginSet:", err)
	}
	if err := cli.PluginEnable(ctx, alias, &client.PluginEnable{Timeout: 30}); err != nil {
		t.Fatalf("PluginEnable after set: %v", err)
	}
	if err := cli.PluginDisable(ctx, alias, &client.PluginDisable{Force: true}); err != nil {
		t.Fatalf("PluginDisable before remove: %v", err)
	}
	if err := cli.PluginRemove(ctx, alias, &client.PluginRemove{Force: true}); err != nil {
		t.Fatalf("PluginRemove: %v", err)
	}
	if _, err := cli.PluginInspect(ctx, alias); !client.IsErrNotFound(err) {
		t.Fatalf("PluginInspect after remove: want not found, got %v", err)
	}
}
