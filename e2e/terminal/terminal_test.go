//go:build e2e

package terminal_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aptd3v/containerkit/pkg/kit"
)

const (
	clientImage = "alpine:3.21"
	marker      = "containerkit-tty-ok"
	attachName  = "containerkit-e2e-tty-attach"
	execName    = "containerkit-e2e-tty-exec"
	stdinName   = "containerkit-e2e-tty-stdin"
	execStdin   = "containerkit-e2e-tty-exec-stdin"
)

func TestAttachTerminal(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)
	api := liveEngine(t, ctx)

	id := createTTY(t, ctx, api, attachName, kit.Command("echo", marker))
	inR, _ := pipeFiles(t)
	var out bytes.Buffer
	sess, err := kit.AttachTerminal(ctx, api, id, inR, &out)
	if err != nil {
		t.Fatalf("AttachTerminal: %v", err)
	}

	done := make(chan error, 1)
	go func() { done <- sess.Start(ctx, nil) }()
	if err := kit.NewContainer().Start(ctx, api, id); err != nil {
		t.Fatalf("Start: %v", err)
	}

	if err := waitSession(ctx, t, done, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), marker) {
		t.Fatalf("output missing %q:\n%s", marker, dump(out.Bytes()))
	}
}

func TestAttachTerminalStdin(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)
	api := liveEngine(t, ctx)

	id := createTTY(t, ctx, api, stdinName, kit.Command("sh", "-c", `IFS= read -r line; printf '%s\n' "$line"`))
	inR, inW := pipeFiles(t)
	out := &safeBuf{}
	sess, err := kit.AttachTerminal(ctx, api, id, inR, out)
	if err != nil {
		t.Fatalf("AttachTerminal: %v", err)
	}

	done := make(chan error, 1)
	go func() { done <- sess.Start(ctx, nil) }()
	waitBrief(ctx, t, 50*time.Millisecond)

	// Bytes must be on the hijack before the process reads; otherwise cat/read
	// can see EOF on a TTY whose attach stdin copier has not started.
	if _, err := io.WriteString(inW, marker+"\n"); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := kit.NewContainer().Start(ctx, api, id); err != nil {
		t.Fatalf("Start: %v", err)
	}

	waitForMarker(ctx, t, api, id, out, done)
}

func TestExecTerminal(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)
	api := liveEngine(t, ctx)

	id := createTTY(t, ctx, api, execName, kit.Command("sleep", "86400"))
	if err := kit.NewContainer().Start(ctx, api, id); err != nil {
		t.Fatalf("Start: %v", err)
	}

	inR, _ := pipeFiles(t)
	var out bytes.Buffer
	sess, err := kit.ExecTerminal(ctx, api, id, inR, &out, kit.Command("echo", marker))
	if err != nil {
		t.Fatalf("ExecTerminal: %v", err)
	}

	done := make(chan error, 1)
	go func() { done <- sess.Start(ctx, nil) }()
	if err := waitSession(ctx, t, done, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), marker) {
		t.Fatalf("output missing %q:\n%s", marker, dump(out.Bytes()))
	}
}

func TestExecTerminalStdin(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)
	api := liveEngine(t, ctx)

	id := createTTY(t, ctx, api, execStdin, kit.Command("sleep", "86400"))
	if err := kit.NewContainer().Start(ctx, api, id); err != nil {
		t.Fatalf("Start: %v", err)
	}

	inR, inW := pipeFiles(t)
	out := &safeBuf{}
	sess, err := kit.ExecTerminal(ctx, api, id, inR, out, kit.Command("sh", "-c", `IFS= read -r line; printf '%s\n' "$line"`))
	if err != nil {
		t.Fatalf("ExecTerminal: %v", err)
	}

	done := make(chan error, 1)
	go func() { done <- sess.Start(ctx, nil) }()
	waitBrief(ctx, t, 50*time.Millisecond)
	if _, err := io.WriteString(inW, marker+"\n"); err != nil {
		t.Fatalf("write: %v", err)
	}

	waitForMarker(ctx, t, api, id, out, done)
}

func liveEngine(t *testing.T, ctx context.Context) kit.Engine {
	t.Helper()
	api, err := kit.NewEngine()
	if err != nil {
		t.Skipf("NewEngine: %v", err)
	}
	t.Cleanup(func() { _ = api.Close() })
	if _, err := kit.Ping(ctx, api); err != nil {
		t.Skipf("engine not reachable: %v", err)
	}
	if _, err := kit.PullImage(ctx, api, clientImage); err != nil {
		t.Skipf("PullImage %s: %v", clientImage, err)
	}
	return api
}

func createTTY(t *testing.T, ctx context.Context, api kit.Engine, name string, cmd kit.ContainerOption) string {
	t.Helper()
	box := kit.NewContainer(
		kit.Name(name),
		kit.Image(clientImage),
		cmd,
		kit.Tty(true),
		kit.Interactive(true),
		kit.StdinOnce(true),
		kit.AttachStdin(true),
		kit.AttachStdout(true),
		kit.AttachStderr(true),
	)
	if err := box.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	_ = box.Remove(ctx, api, name, kit.Force(true))
	created, err := box.Create(ctx, api)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() {
		rmCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
		defer cancel()
		_ = box.Remove(rmCtx, api, created.ID, kit.Force(true))
	})
	return created.ID
}

func waitSession(ctx context.Context, t *testing.T, done <-chan error, out *bytes.Buffer) error {
	t.Helper()
	select {
	case err := <-done:
		if err != nil {
			t.Logf("output:\n%s", dump(out.Bytes()))
		}
		return err
	case <-ctx.Done():
		t.Logf("output:\n%s", dump(out.Bytes()))
		return ctx.Err()
	}
}

func waitForMarker(ctx context.Context, t *testing.T, api kit.Engine, id string, out *safeBuf, done <-chan error) {
	t.Helper()
	tick := time.NewTicker(20 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("timeout waiting for %q: %v\n%s\n%s", marker, ctx.Err(), inspectTTY(ctx, api, id), dump(out.Bytes()))
		case err := <-done:
			got := out.Bytes()
			if !bytes.Contains(got, []byte(marker)) {
				t.Fatalf("session ended (err=%v) missing %q\n%s\n%s", err, marker, inspectTTY(ctx, api, id), dump(got))
			}
			return
		case <-tick.C:
			if strings.Contains(out.String(), marker) {
				return
			}
		}
	}
}

func waitBrief(ctx context.Context, t *testing.T, d time.Duration) {
	t.Helper()
	select {
	case <-time.After(d):
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
}

func inspectTTY(ctx context.Context, api kit.Engine, id string) string {
	ins, err := kit.NewContainer().Inspect(ctx, api, id)
	if err != nil {
		return fmt.Sprintf("inspect: %v", err)
	}
	cfg := ins.Container.Config
	if cfg == nil {
		return "inspect: nil config"
	}
	st := ins.Container.State
	state := ""
	if st != nil {
		state = fmt.Sprintf(" running=%v exit=%d error=%q", st.Running, st.ExitCode, st.Error)
	}
	return fmt.Sprintf("tty=%v openStdin=%v attachStdin=%v stdinOnce=%v cmd=%q%s",
		cfg.Tty, cfg.OpenStdin, cfg.AttachStdin, cfg.StdinOnce, cfg.Cmd, state)
}

func dump(b []byte) string {
	if len(b) == 0 {
		return "output <empty>"
	}
	return fmt.Sprintf("output %q\nhex %s", b, hex.EncodeToString(b))
}

func pipeFiles(t *testing.T) (*os.File, *os.File) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		r.Close()
		w.Close()
	})
	return r, w
}

type safeBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *safeBuf) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *safeBuf) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

func (s *safeBuf) Bytes() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]byte(nil), s.b.Bytes()...)
}
