//go:build e2e

package swarm_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/aptd3v/containerkit/pkg/kit"
)

func TestSwarmFourNodeCluster(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	t.Cleanup(cancel)

	managerHost, nodeHosts, advertiseIPs := clusterHosts(t)
	engines := make([]kit.Engine, 0, len(nodeHosts))
	var pingErrs []string
	for i, host := range nodeHosts {
		api, err := kit.NewEngine(kit.Host(host))
		if err != nil {
			t.Fatalf("NewEngine(%s): %v", host, err)
		}
		t.Cleanup(func() { _ = api.Close() })
		if _, err := kit.Ping(ctx, api); err != nil {
			pingErrs = append(pingErrs, host+": "+err.Error())
			continue
		}
		engines = append(engines, api)
		t.Logf("ping ok %s (%s)", host, advertiseIPs[i])
	}
	if len(engines) == 0 {
		t.Skip("swarm cluster not reachable; start with e2e/swarm/up.sh")
	}
	if len(engines) != 4 {
		t.Fatalf("only %d/4 nodes reachable: %s", len(engines), strings.Join(pingErrs, "; "))
	}

	resetSwarm(ctx, t, engines)

	mgr := engines[0]
	var initRes kit.SwarmInitResult
	var err error
	for attempt := 0; attempt < 5; attempt++ {
		initRes, err = kit.InitSwarm(ctx, mgr,
			kit.ListenAddr("0.0.0.0:2377"),
			kit.AdvertiseAddr(advertiseIPs[0]+":2377"),
		)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		t.Fatalf("InitSwarm: %v", err)
	}
	if initRes.NodeID == "" {
		t.Fatal("InitSwarm returned empty node id")
	}

	info, err := kit.InspectSwarm(ctx, mgr)
	if err != nil {
		t.Fatalf("InspectSwarm: %v", err)
	}
	token := info.Swarm.JoinTokens.Worker
	if token == "" {
		t.Fatal("empty worker join token")
	}

	remote := advertiseIPs[0] + ":2377"
	for i := 1; i < len(engines); i++ {
		if _, err := kit.JoinSwarm(ctx, engines[i],
			kit.ListenAddr("0.0.0.0:2377"),
			kit.AdvertiseAddr(advertiseIPs[i]+":2377"),
			kit.RemoteAddrs(remote),
			kit.JoinToken(token),
		); err != nil {
			t.Fatalf("JoinSwarm %s: %v", nodeHosts[i], err)
		}
	}

	nodes, err := kit.ListNodes(ctx, mgr)
	if err != nil {
		t.Fatalf("ListNodes: %v", err)
	}
	if len(nodes.Items) != 4 {
		t.Fatalf("ListNodes = %d, want 4", len(nodes.Items))
	}
	managers := 0
	for _, n := range nodes.Items {
		if n.Spec.Role == kit.NodeRoleManager {
			managers++
		}
	}
	if managers != 1 {
		t.Fatalf("managers = %d, want 1", managers)
	}

	created, err := kit.CreateSwarmService(ctx, mgr,
		kit.Name("containerkit-e2e-sleep"),
		kit.Image("alpine:3.21"),
		kit.Command("sleep", "infinity"),
	)
	if err != nil {
		t.Fatalf("CreateSwarmService: %v", err)
	}
	if created.ID == "" {
		t.Fatal("CreateSwarmService returned empty id")
	}
	t.Cleanup(func() {
		_, _ = kit.RemoveSwarmService(context.Background(), mgr, created.ID)
	})

	listed, err := kit.ListSwarmServices(ctx, mgr)
	if err != nil {
		t.Fatalf("ListSwarmServices: %v", err)
	}
	found := false
	for _, svc := range listed.Items {
		if svc.ID == created.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created service %s not in ListSwarmServices", created.ID)
	}

	if managerHost == "" {
		t.Fatal("missing manager host")
	}
}

func resetSwarm(ctx context.Context, t *testing.T, engines []kit.Engine) {
	t.Helper()
	for _, api := range engines {
		if _, err := kit.LeaveSwarm(ctx, api, kit.Force(true)); err != nil {
			t.Logf("LeaveSwarm: %v", err)
		}
	}
}

func clusterHosts(t *testing.T) (manager string, nodes []string, ips []string) {
	t.Helper()
	loadRunEnv()
	manager = getenv("CONTAINERKIT_SWARM_MANAGER", "tcp://127.0.0.1:23751")
	nodes = splitCSV(getenv("CONTAINERKIT_SWARM_NODES",
		"tcp://127.0.0.1:23751,tcp://127.0.0.1:23752,tcp://127.0.0.1:23753,tcp://127.0.0.1:23754"))
	ips = splitCSV(getenv("CONTAINERKIT_SWARM_NODE_IPS", "10.0.0.11,10.0.0.12,10.0.0.13,10.0.0.14"))
	if len(nodes) != 4 || len(ips) != 4 {
		t.Fatalf("need 4 nodes and 4 advertise IPs, got %d hosts %d ips", len(nodes), len(ips))
	}
	return manager, nodes, ips
}

func loadRunEnv() {
	candidates := []string{"e2e/swarm/.run/env", ".run/env"}
	if _, file, _, ok := runtime.Caller(0); ok {
		candidates = append([]string{filepath.Join(filepath.Dir(file), ".run/env")}, candidates...)
	}
	for _, p := range candidates {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(b), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			k, v, ok := strings.Cut(line, "=")
			if !ok {
				continue
			}
			k = strings.TrimSpace(k)
			v = strings.TrimSpace(v)
			if k != "" && os.Getenv(k) == "" {
				_ = os.Setenv(k, v)
			}
		}
		return
	}
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
