//go:build e2e

package e2e

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aptd3v/containerkit/pkg/client"
	"github.com/aptd3v/containerkit/pkg/kit"
)

func TestClientMeta(t *testing.T) {
	t.Parallel()
	cli, err := client.NewClient(
		client.FromEnv(),
		client.WithAPIVersionNegotiation(),
		client.WithTimeout(time.Minute),
		client.WithUserAgent("containerkit-e2e"),
		client.WithHTTPHeaders(map[string]string{"X-Containerkit-E2E": "1"}),
		nil,
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = cli.Close() })

	if cli.Unwrap() == nil {
		t.Fatal("Unwrap: nil sdk client")
	}
	if cli.ClientVersion() == "" {
		t.Fatal("ClientVersion: empty")
	}
	if cli.DaemonHost() == "" {
		t.Fatal("DaemonHost: empty")
	}
	fmt.Fprintf(testWriter(t), "version=%s host=%s\n", cli.ClientVersion(), cli.DaemonHost())

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	dial := cli.Dialer()
	if dial == nil {
		t.Fatal("Dialer: nil")
	}
	conn, err := dial(ctx)
	if err != nil {
		t.Fatalf("Dialer: %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("Dialer close: %v", err)
	}

	if hij, err := cli.DialHijack(ctx, "/events", "tcp", nil); err != nil {
		fmt.Fprintln(testWriter(t), "DialHijack:", err)
	} else {
		_ = hij.Close()
	}
}

func TestClientSystem(t *testing.T) {
	t.Parallel()
	cli := newCLI(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	info, err := cli.Info(ctx)
	if err != nil {
		t.Fatalf("Info: %v", err)
	}
	if info == nil || info.ID == "" {
		t.Fatal("Info: empty daemon ID")
	}

	pong, err := cli.Ping(ctx, nil)
	if err != nil {
		t.Fatalf("Ping nil: %v", err)
	}
	if pong == nil || pong.APIVersion == "" {
		t.Fatal("Ping: empty API version")
	}
	if _, err := cli.Ping(ctx, &client.Ping{NegotiateAPIVersion: true, ForceNegotiate: true}); err != nil {
		t.Fatalf("Ping negotiate: %v", err)
	}

	ver, err := cli.ServerVersion(ctx)
	if err != nil {
		t.Fatalf("ServerVersion: %v", err)
	}
	if ver == nil || ver.Version == "" || ver.APIVersion == "" {
		t.Fatal("ServerVersion: empty")
	}

	usage, err := cli.DiskUsage(ctx, nil)
	if err != nil {
		t.Fatalf("DiskUsage nil: %v", err)
	}
	if usage == nil {
		t.Fatal("DiskUsage: nil")
	}
	if _, err := cli.DiskUsage(ctx, &client.DiskUsage{Containers: true, Images: true, Verbose: true}); err != nil {
		t.Fatalf("DiskUsage subset: %v", err)
	}
	if _, err := cli.DiskUsage(ctx, &client.DiskUsage{Volumes: true, BuildCache: true}); err != nil {
		t.Fatalf("DiskUsage volumes: %v", err)
	}

	if dist, err := cli.DistributionInspect(ctx, alpine, nil); err != nil {
		fmt.Fprintln(testWriter(t), "DistributionInspect:", err)
	} else if dist == nil {
		t.Fatal("DistributionInspect: nil")
	}
	if _, err := cli.DistributionInspect(ctx, alpine, &client.DistributionInspect{}); err != nil {
		fmt.Fprintln(testWriter(t), "DistributionInspect empty auth:", err)
	}
}

func TestClientEvents(t *testing.T) {
	t.Parallel()
	cli := newCLI(t)
	pullAlpine(t, cli)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	watchCtx, watchCancel := context.WithCancel(ctx)
	msgs, errs := cli.Events(watchCtx, nil)
	watchCancel()
	select {
	case <-msgs:
	case <-errs:
	case <-time.After(2 * time.Second):
	}

	name := ident(t, "c")
	ctr := kit.NewContainer(name).Image(alpine).Command("true")
	created, err := cli.ContainerCreate(ctx, ctr)
	if err != nil {
		t.Fatalf("ContainerCreate: %v", err)
	}
	removeContainer(t, cli, created.ID)

	evCtx, evCancel := context.WithTimeout(ctx, 30*time.Second)
	defer evCancel()
	msgs, errs = cli.Events(evCtx, &client.EngineEvents{
		Filters: make(client.Filters).Add("type", "container").Add("container", created.ID).Add("event", "start"),
	})
	if err := cli.ContainerStart(ctx, created.ID, nil); err != nil {
		t.Fatalf("ContainerStart: %v", err)
	}
	select {
	case ev := <-msgs:
		fmt.Fprintln(testWriter(t), "event", ev.Action, ev.Actor.ID)
	case err := <-errs:
		if err != nil {
			t.Fatalf("Events: %v", err)
		}
	case <-evCtx.Done():
		t.Fatal("Events: timeout waiting for container start")
	}
}

func TestClientEngineErrors(t *testing.T) {
	t.Parallel()
	cli := newCLI(t)
	pullAlpine(t, cli)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	missing := "containerkit-e2e-missing"
	if _, err := cli.ContainerInspect(ctx, missing); !client.IsErrNotFound(err) {
		t.Fatalf("ContainerInspect missing: want not found, got %v", err)
	}
	if _, err := cli.ImageInspect(ctx, missing, nil); !client.IsErrNotFound(err) {
		t.Fatalf("ImageInspect missing: want not found, got %v", err)
	}
	if _, err := cli.NetworkInspect(ctx, missing, nil); !client.IsErrNotFound(err) {
		t.Fatalf("NetworkInspect missing: want not found, got %v", err)
	}
	if _, err := cli.VolumeInspect(ctx, missing); !client.IsErrNotFound(err) {
		t.Fatalf("VolumeInspect missing: want not found, got %v", err)
	}

	name := ident(t, "c")
	ctr := kit.NewContainer(name).Image(alpine).Command("true")
	created, err := cli.ContainerCreate(ctx, ctr)
	if err != nil {
		t.Fatalf("ContainerCreate: %v", err)
	}
	removeContainer(t, cli, created.ID)
	if _, err := cli.ContainerCreate(ctx, ctr); !client.IsErrConflict(err) {
		t.Fatalf("ContainerCreate dup: want conflict, got %v", err)
	}

	canceled, cancelNow := context.WithCancel(context.Background())
	cancelNow()
	if _, err := cli.Info(canceled); err == nil {
		t.Fatal("Info canceled: want error")
	} else if !client.IsErrCanceled(err) && !errors.Is(err, context.Canceled) {
		fmt.Fprintln(testWriter(t), "Info canceled:", err)
	}

	deadline, expire := context.WithTimeout(context.Background(), time.Nanosecond)
	defer expire()
	time.Sleep(2 * time.Millisecond)
	if _, err := cli.ServerVersion(deadline); err == nil {
		t.Fatal("ServerVersion deadline: want error")
	} else if !client.IsErrDeadlineExceeded(err) && !errors.Is(err, context.DeadlineExceeded) {
		fmt.Fprintln(testWriter(t), "ServerVersion deadline:", err)
	}

	if err := cli.ContainerKill(ctx, created.ID, "NOTASIGNAL"); err != nil && !client.IsErrInvalidArgument(err) && !client.IsErrNotFound(err) {
		fmt.Fprintln(testWriter(t), "ContainerKill invalid:", err)
	}
}

func TestClientNetworkAndImageFilters(t *testing.T) {
	t.Parallel()
	cli := newCLI(t)
	pullAlpine(t, cli)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	netName := ident(t, "n")
	net, err := cli.NetworkCreate(ctx, netName, &client.NetworkCreate{
		Driver: "bridge",
		Labels: map[string]string{"containerkit-e2e": "1"},
	})
	if err != nil {
		t.Fatalf("NetworkCreate: %v", err)
	}
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		_ = cli.NetworkRemove(c, net.ID)
	})
	if _, err := cli.NetworkInspect(ctx, net.ID, &client.NetworkInspect{Verbose: true}); err != nil {
		t.Fatalf("NetworkInspect verbose: %v", err)
	}
	listed, err := cli.NetworkList(ctx, &client.NetworkList{Filters: make(client.Filters).Add("name", netName)})
	if err != nil {
		t.Fatalf("NetworkList filter: %v", err)
	}
	found := false
	for _, n := range listed {
		if n.ID == net.ID || n.Name == netName {
			found = true
		}
	}
	if !found {
		t.Fatal("NetworkList filter: created network missing")
	}

	imgs, err := cli.ImageList(ctx, &client.ImageList{
		All:        true,
		SharedSize: true,
		Filters:    make(client.Filters).Add("reference", alpine),
	})
	if err != nil {
		t.Fatalf("ImageList filter: %v", err)
	}
	found = false
	for _, img := range imgs {
		for _, tag := range img.RepoTags {
			if tag == alpine || strings.HasPrefix(tag, "alpine:") {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("ImageList filter: alpine missing")
	}

	vol, err := cli.VolumeCreate(ctx, &client.VolumeCreate{
		Name:   ident(t, "v"),
		Labels: map[string]string{"containerkit-e2e": "1"},
	})
	if err != nil {
		t.Fatalf("VolumeCreate: %v", err)
	}
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		_ = cli.VolumeRemove(c, vol.Name, true)
	})
	vols, err := cli.VolumeList(ctx, &client.VolumeList{Filters: make(client.Filters).Add("name", vol.Name)})
	if err != nil {
		t.Fatalf("VolumeList filter: %v", err)
	}
	found = false
	if vols != nil {
		for _, v := range vols.Volumes {
			if v.Name == vol.Name {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("VolumeList filter: created volume missing")
	}
	_ = io.Discard
}
