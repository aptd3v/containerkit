//go:build e2e

package e2e

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aptd3v/containerkit/pkg/client"
	"github.com/aptd3v/containerkit/pkg/containerkit"
)

func TestClientImagePullInspectList(t *testing.T) {
	t.Parallel()
	cli := newCLI(t)
	pullAlpine(t, cli)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	insp, err := cli.ImageInspect(ctx, alpine)
	if err != nil {
		t.Fatalf("ImageInspect: %v", err)
	}
	if insp.ID == "" {
		t.Fatal("ImageInspect: empty ID")
	}

	list, err := cli.ImageList(ctx, &client.ImageList{})
	if err != nil {
		t.Fatalf("ImageList: %v", err)
	}
	found := false
	for _, img := range list {
		for _, tag := range img.RepoTags {
			if tag == alpine || strings.HasPrefix(tag, "alpine:") {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("ImageList: alpine:latest not found")
	}
}

func TestClientContainerLifecycle(t *testing.T) {
	t.Parallel()
	cli := newCLI(t)
	pullAlpine(t, cli)
	name := ident(t, "c")
	ctr := containerkit.NewContainer(name).
		Image(alpine).
		Command("sleep", "infinity")
	if err := ctr.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	created, err := cli.ContainerCreate(ctx, ctr)
	if err != nil {
		t.Fatalf("ContainerCreate: %v", err)
	}
	id := created.ID
	if id == "" {
		id = name
	}
	removeContainer(t, cli, id)

	if err := cli.ContainerStart(ctx, id, nil); err != nil {
		t.Fatalf("ContainerStart nil opts: %v", err)
	}

	insp, err := cli.ContainerInspect(ctx, id)
	if err != nil {
		t.Fatalf("ContainerInspect: %v", err)
	}
	if insp.State == nil || !insp.State.Running {
		t.Fatal("container not running")
	}

	logs, err := cli.ContainerLogs(ctx, id, &client.Logs{ShowStdout: true, ShowStderr: true, Tail: "10"})
	if err != nil {
		t.Fatalf("ContainerLogs: %v", err)
	}
	_, _ = io.Copy(testWriter(t), logs)
	logs.Close()

	execResp, err := cli.ContainerExecCreate(ctx, id, &client.Exec{
		AttachStdout: true,
		AttachStderr: true,
		Command:      []string{"echo", "client-ok"},
	})
	if err != nil {
		t.Fatalf("ContainerExecCreate: %v", err)
	}
	hij, err := cli.ContainerExecAttach(ctx, execResp.ID, &client.ExecAttach{})
	if err != nil {
		t.Fatalf("ContainerExecAttach: %v", err)
	}
	out := drainExec(t, hij.Reader)
	hij.Close()
	if !strings.Contains(out, "client-ok") {
		t.Fatalf("exec output %q, want client-ok", out)
	}

	execInsp, err := cli.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		t.Fatalf("ContainerExecInspect: %v", err)
	}
	if execInsp.ExitCode != 0 {
		t.Fatalf("exec exit code %d", execInsp.ExitCode)
	}

	if _, err := cli.ContainerTop(ctx, id); err != nil {
		t.Fatalf("ContainerTop: %v", err)
	}
	stats, err := cli.ContainerStatsOneShot(ctx, id)
	if err != nil {
		t.Fatalf("ContainerStatsOneShot: %v", err)
	}
	if stats != nil && stats.Body != nil {
		_, _ = io.Copy(testWriter(t), stats.Body)
		stats.Body.Close()
	}

	if err := cli.ContainerPause(ctx, id); err != nil {
		t.Fatalf("ContainerPause: %v", err)
	}
	insp, err = cli.ContainerInspect(ctx, id)
	if err != nil {
		t.Fatalf("Inspect after pause: %v", err)
	}
	if insp.State == nil || !insp.State.Paused {
		t.Fatal("container not paused")
	}
	if err := cli.ContainerUnpause(ctx, id); err != nil {
		t.Fatalf("ContainerUnpause: %v", err)
	}

	if err := cli.ContainerStop(ctx, id, &client.Stop{}); err != nil {
		t.Fatalf("ContainerStop: %v", err)
	}
	if err := cli.ContainerRemove(ctx, id, &client.Remove{Force: true, Volumes: true}); err != nil {
		t.Fatalf("ContainerRemove: %v", err)
	}
}

func TestClientNetwork(t *testing.T) {
	t.Parallel()
	cli := newCLI(t)
	pullAlpine(t, cli)
	netName := ident(t, "n")
	ctrName := ident(t, "c")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	net, err := cli.NetworkCreate(ctx, netName, &client.NetworkCreate{Driver: "bridge"})
	if err != nil {
		t.Fatalf("NetworkCreate: %v", err)
	}
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		_ = cli.NetworkRemove(c, net.ID)
	})

	if _, err := cli.NetworkInspect(ctx, net.ID, nil); err != nil {
		t.Fatalf("NetworkInspect nil opts: %v", err)
	}
	nets, err := cli.NetworkList(ctx, nil)
	if err != nil {
		t.Fatalf("NetworkList: %v", err)
	}
	found := false
	for _, n := range nets {
		if n.ID == net.ID || n.Name == netName {
			found = true
		}
	}
	if !found {
		t.Fatal("NetworkList: created network missing")
	}

	ctr := containerkit.NewContainer(ctrName).Image(alpine).Command("sleep", "infinity")
	created, err := cli.ContainerCreate(ctx, ctr)
	if err != nil {
		t.Fatalf("ContainerCreate: %v", err)
	}
	removeContainer(t, cli, created.ID)
	if err := cli.ContainerStart(ctx, created.ID, nil); err != nil {
		t.Fatalf("ContainerStart: %v", err)
	}

	if err := cli.NetworkConnect(ctx, net.ID, &client.NetworkConnect{Container: created.ID}); err != nil {
		t.Fatalf("NetworkConnect: %v", err)
	}
	if err := cli.NetworkDisconnect(ctx, net.ID, created.ID, true); err != nil {
		t.Fatalf("NetworkDisconnect: %v", err)
	}
	if err := cli.ContainerRemove(ctx, created.ID, &client.Remove{Force: true}); err != nil {
		t.Fatalf("ContainerRemove: %v", err)
	}
	if err := cli.NetworkRemove(ctx, net.ID); err != nil {
		t.Fatalf("NetworkRemove: %v", err)
	}
}

func TestClientVolume(t *testing.T) {
	t.Parallel()
	cli := newCLI(t)
	volName := ident(t, "v")
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	vol, err := cli.VolumeCreate(ctx, &client.VolumeCreate{Name: volName})
	if err != nil {
		t.Fatalf("VolumeCreate: %v", err)
	}
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		_ = cli.VolumeRemove(c, vol.Name, true)
	})

	if _, err := cli.VolumeInspect(ctx, vol.Name); err != nil {
		t.Fatalf("VolumeInspect: %v", err)
	}
	list, err := cli.VolumeList(ctx, nil)
	if err != nil {
		t.Fatalf("VolumeList: %v", err)
	}
	found := false
	if list != nil {
		for _, v := range list.Volumes {
			if v.Name == vol.Name {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatal("VolumeList: created volume missing")
	}
	if err := cli.VolumeRemove(ctx, vol.Name, true); err != nil {
		t.Fatalf("VolumeRemove: %v", err)
	}
}

func TestClientImageBuild(t *testing.T) {
	t.Parallel()
	cli := newCLI(t)
	pullAlpine(t, cli)
	tag := ident(t, "i") + ":e2e"
	df := containerkit.NewDockerFile().From("alpine", "latest").CommandExec("true")
	ctxReader, err := df.NewLocalBuildContext(t.TempDir())
	if err != nil {
		t.Fatalf("NewLocalBuildContext: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	opt := &client.ImageBuild{}
	opt.Tags = []string{tag}
	opt.Remove = true
	resp, err := cli.ImageBuild(ctx, ctxReader, opt)
	if err != nil {
		t.Fatalf("ImageBuild: %v", err)
	}
	drainBuild(t, resp.Body)

	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = cli.ImageRemove(c, tag, &client.ImageRemove{Force: true, PruneChildren: true})
	})

	insp, err := cli.ImageInspect(ctx, tag)
	if err != nil {
		t.Fatalf("ImageInspect built: %v", err)
	}
	if insp.ID == "" {
		t.Fatal("built image empty ID")
	}
	if _, err := cli.ImageRemove(ctx, tag, &client.ImageRemove{Force: true, PruneChildren: true}); err != nil {
		t.Fatalf("ImageRemove: %v", err)
	}
}

func TestClientNilOptions(t *testing.T) {
	t.Parallel()
	cli := newCLI(t)
	pullAlpine(t, cli)
	name := ident(t, "c")
	ctr := containerkit.NewContainer(name).Image(alpine).Command("true")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	created, err := cli.ContainerCreate(ctx, ctr)
	if err != nil {
		t.Fatalf("ContainerCreate: %v", err)
	}
	removeContainer(t, cli, created.ID)

	if err := cli.ContainerStart(ctx, created.ID, nil); err != nil {
		t.Fatalf("ContainerStart nil: %v", err)
	}
	_ = cli.ContainerWaitSync(ctx, created.ID, client.WaitConditionNotRunning)
	if err := cli.ContainerRemove(ctx, created.ID, nil); err != nil {
		t.Fatalf("ContainerRemove nil: %v", err)
	}

	if _, err := cli.ImageList(ctx, nil); err != nil {
		t.Fatalf("ImageList nil: %v", err)
	}
	if _, err := cli.NetworkList(ctx, nil); err != nil {
		t.Fatalf("NetworkList nil: %v", err)
	}
	if _, err := cli.VolumeList(ctx, nil); err != nil {
		t.Fatalf("VolumeList nil: %v", err)
	}
}

func TestClientSwarmInspect(t *testing.T) {
	t.Parallel()
	cli := newCLI(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, err := cli.SwarmInspect(ctx)
	if err != nil {
		fmt.Fprintln(testWriter(t), err.Error())
		if err.Error() == "" {
			t.Fatal("SwarmInspect error with empty message")
		}
	}
	if err := cli.SwarmJoin(ctx, &client.SwarmJoin{RemoteAddrs: []string{"127.0.0.1:2377"}, JoinToken: "e2e"}); err != nil {
		fmt.Fprintln(testWriter(t), "SwarmJoin:", err)
	}
	if err := cli.SwarmLeave(ctx, false); err != nil {
		fmt.Fprintln(testWriter(t), "SwarmLeave:", err)
	}
}

func TestClientContainerMoreOps(t *testing.T) {
	t.Parallel()
	cli := newCLI(t)
	pullAlpine(t, cli)
	name := ident(t, "c")
	renamed := ident(t, "r")
	ctr := containerkit.NewContainer(name).Image(alpine).Command("sleep", "infinity")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	created, err := cli.ContainerCreate(ctx, ctr)
	if err != nil {
		t.Fatalf("ContainerCreate: %v", err)
	}
	id := created.ID
	removeContainer(t, cli, id)
	if err := cli.ContainerStart(ctx, id, &client.Start{}); err != nil {
		t.Fatalf("ContainerStart: %v", err)
	}

	list, err := cli.ContainerList(ctx, &client.List{All: true, Size: true, Filters: []client.Filter{{Key: "name", Value: name}}})
	if err != nil {
		t.Fatalf("ContainerList: %v", err)
	}
	found := false
	for _, s := range list {
		if s.ID == id || strings.Contains(strings.Join(s.Names, " "), name) {
			found = true
		}
	}
	if !found {
		t.Fatal("ContainerList: created container missing")
	}

	if err := cli.ContainerRename(ctx, id, renamed); err != nil {
		t.Fatalf("ContainerRename: %v", err)
	}
	if _, err := cli.ContainerUpdate(ctx, id, &client.Update{CPUShares: 512, RestartPolicy: containerkit.RestartPolicyOnFailure, RestartMaxRetry: 1}); err != nil {
		t.Fatalf("ContainerUpdate: %v", err)
	}

	var tarbuf bytes.Buffer
	tw := tar.NewWriter(&tarbuf)
	body := []byte("hello")
	if err := tw.WriteHeader(&tar.Header{Name: "hello.txt", Size: int64(len(body)), Mode: 0644}); err != nil {
		t.Fatalf("tar header: %v", err)
	}
	if _, err := tw.Write(body); err != nil {
		t.Fatalf("tar write: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar close: %v", err)
	}
	if err := cli.ContainerCopyToContainer(ctx, id, "/tmp", &client.CopyTo{Content: &tarbuf}); err != nil {
		t.Fatalf("CopyToContainer: %v", err)
	}
	rc, stat, err := cli.ContainerCopyFromContainer(ctx, id, "/tmp/hello.txt")
	if err != nil {
		t.Fatalf("CopyFromContainer: %v", err)
	}
	if stat == nil {
		t.Fatal("CopyFromContainer: nil stat")
	}
	_, _ = io.Copy(io.Discard, rc)
	rc.Close()

	if _, err := cli.ContainerDiff(ctx, id); err != nil {
		t.Fatalf("ContainerDiff: %v", err)
	}
	exp, err := cli.ContainerExport(ctx, id)
	if err != nil {
		t.Fatalf("ContainerExport: %v", err)
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(exp, 1024))
	exp.Close()

	stats, err := cli.ContainerStats(ctx, id, false)
	if err != nil {
		t.Fatalf("ContainerStats: %v", err)
	}
	if stats != nil && stats.Body != nil {
		_, _ = io.Copy(testWriter(t), io.LimitReader(stats.Body, 4096))
		stats.Body.Close()
	}

	if err := cli.ContainerResize(ctx, id, &client.Resize{Width: 80, Height: 24}); err != nil {
		fmt.Fprintln(testWriter(t), "ContainerResize:", err)
	}

	execResp, err := cli.ContainerExecCreate(ctx, id, &client.Exec{
		Detach:  true,
		Command: []string{"true"},
	})
	if err != nil {
		t.Fatalf("ContainerExecCreate detach: %v", err)
	}
	if err := cli.ContainerExecStart(ctx, execResp.ID, &client.ExecStart{Detach: true}); err != nil {
		t.Fatalf("ContainerExecStart: %v", err)
	}
	if err := cli.ContainerExecResize(ctx, execResp.ID, &client.Resize{Width: 80, Height: 24}); err != nil {
		fmt.Fprintln(testWriter(t), "ContainerExecResize:", err)
	}

	att, err := cli.ContainerAttach(ctx, id, &client.Attach{Logs: true, Stdout: true, Stderr: true})
	if err != nil {
		t.Fatalf("ContainerAttach: %v", err)
	}
	att.Close()

	if _, err := cli.ContainerExecAttachTerminal(ctx, execResp.ID, &client.ExecAttach{}); err != nil {
		fmt.Fprintln(testWriter(t), "ContainerExecAttachTerminal:", err)
	}

	_ = cli.ContainerCheckpointCreate(ctx, id, &client.CheckpointCreate{CheckpointID: "e2e"})
	if _, err := cli.ContainerCheckpointList(ctx, id, nil); err != nil {
		fmt.Fprintln(testWriter(t), "CheckpointList:", err)
	}
	_ = cli.ContainerCheckpointDelete(ctx, id, &client.CheckpointDelete{CheckpointID: "e2e"})

	tag := ident(t, "i") + ":commit"
	commit, err := cli.ContainerCommit(ctx, id, &client.Commit{Reference: tag, Comment: "e2e", Pause: false})
	if err != nil {
		t.Fatalf("ContainerCommit: %v", err)
	}
	fmt.Fprintln(testWriter(t), "commit", commit.ID)
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		_, _ = cli.ImageRemove(c, tag, &client.ImageRemove{Force: true, PruneChildren: true})
	})

	if err := cli.ContainerRestart(ctx, id, &client.Stop{Signal: "SIGTERM"}); err != nil {
		t.Fatalf("ContainerRestart: %v", err)
	}
	if err := cli.ContainerKill(ctx, id, "SIGKILL"); err != nil {
		t.Fatalf("ContainerKill: %v", err)
	}

	never := []client.Filter{{Key: "label", Value: "containerkit-e2e=never"}}
	if _, err := cli.ContainerPrune(ctx, &client.Prune{Filters: never}); err != nil {
		t.Fatalf("ContainerPrune: %v", err)
	}
}

func TestClientImageMoreOps(t *testing.T) {
	t.Parallel()
	cli := newCLI(t)
	pullAlpine(t, cli)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	if _, err := cli.ImageHistory(ctx, alpine); err != nil {
		t.Fatalf("ImageHistory: %v", err)
	}
	tag := ident(t, "i") + ":tag"
	if err := cli.ImageTag(ctx, alpine, tag); err != nil {
		t.Fatalf("ImageTag: %v", err)
	}
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		_, _ = cli.ImageRemove(c, tag, &client.ImageRemove{Force: true, PruneChildren: true})
	})

	saved, err := cli.ImageSave(ctx, &client.ImageSave{ImageIDs: []string{tag}})
	if err != nil {
		t.Fatalf("ImageSave: %v", err)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, saved); err != nil {
		saved.Close()
		t.Fatalf("ImageSave copy: %v", err)
	}
	saved.Close()

	load, err := cli.ImageLoad(ctx, &client.ImageLoad{Input: bytes.NewReader(buf.Bytes()), Quiet: true})
	if err != nil {
		t.Fatalf("ImageLoad: %v", err)
	}
	if load != nil && load.Body != nil {
		_, _ = io.Copy(testWriter(t), load.Body)
		load.Body.Close()
	}

	created, err := cli.ImageCreate(ctx, alpine, &client.ImageCreate{})
	if err != nil {
		t.Fatalf("ImageCreate: %v", err)
	}
	_, _ = io.Copy(testWriter(t), created)
	created.Close()

	if _, err := cli.ImageSearch(ctx, "alpine", &client.ImageSearch{Limit: 1}); err != nil {
		fmt.Fprintln(testWriter(t), "ImageSearch:", err)
	}
	if rc, err := cli.ImagePush(ctx, tag); err != nil {
		fmt.Fprintln(testWriter(t), "ImagePush:", err)
	} else {
		_, _ = io.Copy(io.Discard, io.LimitReader(rc, 1024))
		rc.Close()
	}
	if rc, err := cli.ImageImport(ctx, "e2e-import", &client.ImageImport{SourceName: "alpine:latest", Message: "e2e"}); err != nil {
		fmt.Fprintln(testWriter(t), "ImageImport:", err)
	} else {
		_, _ = io.Copy(io.Discard, io.LimitReader(rc, 1024))
		rc.Close()
	}
	if _, err := cli.ImagesPrune(ctx, &client.ImagePrune{Filters: []client.Filter{{Key: "label", Value: "containerkit-e2e=never"}}}); err != nil {
		t.Fatalf("ImagesPrune: %v", err)
	}
	if _, err := cli.ImageSave(ctx, nil); err != nil {
		fmt.Fprintln(testWriter(t), "ImageSave nil:", err)
	}

	if _, err := cli.ImageRemove(ctx, tag, nil); err != nil {
		t.Fatalf("ImageRemove: %v", err)
	}
}

func TestClientNetworkVolumePrune(t *testing.T) {
	t.Parallel()
	cli := newCLI(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	never := []client.Filter{{Key: "label", Value: "containerkit-e2e=never"}}
	if _, err := cli.NetworksPrune(ctx, &client.NetworkPrune{Filters: never}); err != nil {
		t.Fatalf("NetworksPrune: %v", err)
	}
	if _, err := cli.VolumesPrune(ctx, &client.VolumePrune{Filters: never}); err != nil {
		t.Fatalf("VolumesPrune: %v", err)
	}

	vol, err := cli.VolumeCreate(ctx, &client.VolumeCreate{Name: ident(t, "v"), Labels: map[string]string{"e2e": "1"}})
	if err != nil {
		t.Fatalf("VolumeCreate: %v", err)
	}
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		_ = cli.VolumeRemove(c, vol.Name, true)
	})
	if _, raw, err := cli.VolumeInspectWithRaw(ctx, vol.Name); err != nil {
		t.Fatalf("VolumeInspectWithRaw: %v", err)
	} else if len(raw) == 0 {
		t.Fatal("VolumeInspectWithRaw: empty raw")
	}
	if err := cli.VolumeUpdate(ctx, vol.Name, 0, nil); err != nil {
		fmt.Fprintln(testWriter(t), "VolumeUpdate:", err)
	}
}
