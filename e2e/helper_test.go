//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode"

	"github.com/aptd3v/containerkit/pkg/client"
	"github.com/aptd3v/containerkit/pkg/kit"
	"github.com/docker/docker/pkg/stdcopy"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

const alpine = "alpine:latest"

func TestMain(m *testing.M) {
	if err := exec.Command("docker", "info").Run(); err != nil {
		fmt.Println("skipping e2e: docker not available:", err)
		os.Exit(0)
	}
	os.Exit(m.Run())
}

type tLogWriter struct {
	t *testing.T
}

func (w *tLogWriter) Write(p []byte) (int, error) {
	s := strings.TrimRight(string(p), "\n")
	if s != "" {
		w.t.Log(s)
	}
	return len(p), nil
}

var testWriters sync.Map // test name -> io.Writer

func logDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "logs"
	}
	return filepath.Join(filepath.Dir(file), "logs")
}

func testWriter(t *testing.T) io.Writer {
	t.Helper()
	if w, ok := testWriters.Load(t.Name()); ok {
		return w.(io.Writer)
	}
	if err := os.MkdirAll(logDir(), 0755); err != nil {
		t.Fatalf("mkdir logs: %v", err)
	}
	name := strings.ReplaceAll(t.Name(), "/", "_")
	path := filepath.Join(logDir(), name+".log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatalf("open log %s: %v", path, err)
	}
	w := io.MultiWriter(f, &tLogWriter{t: t})
	fmt.Fprintf(w, "=== %s ===\n", t.Name())
	t.Logf("e2e log file: %s", path)
	testWriters.Store(t.Name(), w)
	return w
}

func ident(t *testing.T, kind string) string {
	t.Helper()
	n := strings.Map(func(r rune) rune {
		r = unicode.ToLower(r)
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, t.Name())
	n = strings.Trim(n, "-")
	for strings.Contains(n, "--") {
		n = strings.ReplaceAll(n, "--", "-")
	}
	if len(n) > 20 {
		n = n[:20]
	}
	return fmt.Sprintf("e2e%s-%s-%d", kind, n, time.Now().UnixNano()%1e8)
}

func newCLI(t *testing.T) *client.Client {
	t.Helper()
	cli, err := client.NewClient(client.FromEnv(), client.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return cli
}

func pullAlpine(t *testing.T, cli *client.Client) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	rc, err := cli.ImagePull(ctx, alpine, &client.ImagePull{Platforms: []ocispec.Platform{{OS: "linux", Architecture: runtime.GOARCH}}})
	if err != nil {
		t.Fatalf("ImagePull: %v", err)
	}
	defer rc.Close()
	if _, err := io.Copy(testWriter(t), rc); err != nil {
		t.Fatalf("ImagePull stream: %v", err)
	}
}

func drainExec(t *testing.T, r io.Reader) string {
	t.Helper()
	var stdout, stderr strings.Builder
	if _, err := stdcopy.StdCopy(&stdout, &stderr, r); err != nil && err != io.EOF {
		t.Fatalf("stdcopy: %v", err)
	}
	out := stdout.String() + stderr.String()
	if out != "" {
		fmt.Fprint(testWriter(t), out)
	}
	return out
}

func drainBuild(t *testing.T, body io.ReadCloser) {
	t.Helper()
	defer body.Close()
	w := testWriter(t)
	dec := json.NewDecoder(body)
	for {
		var m map[string]any
		if err := dec.Decode(&m); err != nil {
			if err == io.EOF {
				return
			}
			t.Fatalf("ImageBuild stream: %v", err)
		}
		if e, ok := m["error"].(string); ok && e != "" {
			t.Fatalf("ImageBuild: %s", e)
		}
		if s, ok := m["stream"].(string); ok {
			fmt.Fprint(w, s)
		}
		if s, ok := m["status"].(string); ok {
			fmt.Fprintln(w, s)
		}
	}
}

func composeDown(t *testing.T, app *kit.Compose, profiles ...string) {
	t.Helper()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		_ = app.Down(ctx, &kit.Down{
			RemoveOrphans: true,
			RemoveVolumes: true,
			Writer:        testWriter(t),
			Profiles:      profiles,
		})
	})
}

func removeContainer(t *testing.T, cli *client.Client, id string) {
	t.Helper()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = cli.ContainerRemove(ctx, id, &client.Remove{Force: true, RemoveVolumes: true})
	})
}

func removeImage(t *testing.T, cli *client.Client, ref string) {
	t.Helper()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = cli.ImageRemove(ctx, ref, &client.ImageRemove{Force: true, PruneChildren: true})
	})
}

// ensureSwarm makes sure the daemon is in swarm mode. It is not safe to call
// from a parallel test. If this test initialized the swarm, Leave runs on cleanup.
func ensureSwarm(t *testing.T, cli *client.Client) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	_, inspectErr := cli.SwarmInspect(ctx)
	if inspectErr == nil {
		return
	}
	nodeID, initErr := cli.SwarmInit(ctx, &client.SwarmInit{
		ListenAddr:    "0.0.0.0:2377",
		AdvertiseAddr: "127.0.0.1:2377",
	})
	if initErr != nil {
		nodeID, initErr = cli.SwarmInit(ctx, &client.SwarmInit{ListenAddr: "0.0.0.0:2377"})
	}
	if initErr != nil {
		t.Skipf("swarm unavailable: inspect=%v init=%v", inspectErr, initErr)
	}
	fmt.Fprintln(testWriter(t), "SwarmInit", nodeID)
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := cli.SwarmLeave(c, true); err != nil {
			t.Logf("SwarmLeave: %v", err)
		}
	})
}
