package main

import (
	"strings"
	"testing"

	composetypes "github.com/aptd3v/containerkit/internal/compose/types"
)

func TestStackReplacesComposeFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	base := stackConfig{dir: dir, env: map[string]string{}, gateway: "envoy", postgres: "17", pooler: "supavisor", storage: "file"}

	cases := []struct {
		name    string
		mutate  func(stackConfig) stackConfig
		present []string
		absent  []string
		image   map[string]string
	}{
		{
			name:    "envoy postgres 17 supavisor",
			present: []string{"db", "api-gw", "supavisor", "storage", "functions"},
			absent:  []string{"caddy", "nginx", "pgbouncer", "minio", "rustfs", "analytics", "vector"},
			image:   map[string]string{"api-gw": "envoyproxy/envoy:v1.39.6", "db": "supabase/postgres:17.6.1.136"},
		},
		{
			name:    "kong",
			mutate:  func(c stackConfig) stackConfig { c.gateway = "kong"; return c },
			present: []string{"api-gw"},
			image:   map[string]string{"api-gw": "kong/kong:3.9.3"},
		},
		{
			name:    "caddy",
			mutate:  func(c stackConfig) stackConfig { c.gateway = "caddy"; return c },
			present: []string{"api-gw", "caddy"},
			absent:  []string{"nginx"},
			image:   map[string]string{"api-gw": "envoyproxy/envoy:v1.39.6", "caddy": "caddy:2"},
		},
		{
			name:    "nginx",
			mutate:  func(c stackConfig) stackConfig { c.gateway = "nginx"; return c },
			present: []string{"nginx"},
			absent:  []string{"caddy"},
		},
		{
			name:   "postgres 15",
			mutate: func(c stackConfig) stackConfig { c.postgres = "15"; return c },
			image:  map[string]string{"db": "supabase/postgres:15.8.1.085"},
		},
		{
			name:    "pgbouncer",
			mutate:  func(c stackConfig) stackConfig { c.pooler = "pgbouncer"; return c },
			present: []string{"pgbouncer"},
			absent:  []string{"supavisor"},
			image:   map[string]string{"pgbouncer": "edoburu/pgbouncer:v1.25.2-p0"},
		},
		{
			name:    "s3",
			mutate:  func(c stackConfig) stackConfig { c.storage = "s3"; return c },
			present: []string{"minio", "minio-createbucket", "storage"},
			absent:  []string{"rustfs"},
		},
		{
			name:    "rustfs",
			mutate:  func(c stackConfig) stackConfig { c.storage = "rustfs"; return c },
			present: []string{"rustfs", "rustfs-createbucket"},
			absent:  []string{"minio"},
		},
		{
			name:    "logs",
			mutate:  func(c stackConfig) stackConfig { c.logs = true; return c },
			present: []string{"analytics", "vector"},
			image:   map[string]string{"analytics": "supabase/logflare:1.50.15", "vector": "timberio/vector:0.53.0-alpine"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := base
			if tc.mutate != nil {
				cfg = tc.mutate(cfg)
			}
			project, err := setupStack(cfg)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := project.Project(); err != nil {
				t.Fatal(err)
			}
			got := map[string]string{}
			for _, svc := range project.Services() {
				got[svc.Name()] = svc.Image()
			}
			for _, name := range tc.present {
				if _, ok := got[name]; !ok {
					t.Errorf("missing service %s", name)
				}
			}
			for _, name := range tc.absent {
				if _, ok := got[name]; ok {
					t.Errorf("service %s should be absent", name)
				}
			}
			for name, image := range tc.image {
				if got[name] != image {
					t.Errorf("service %s image = %q, want %q", name, got[name], image)
				}
			}
		})
	}
}

func TestStackWiring(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	base := stackConfig{dir: dir, env: map[string]string{"API_GW_HTTP_PORT": "8000", "KONG_HTTPS_PORT": "8443"}, gateway: "envoy", postgres: "17", pooler: "supavisor", storage: "file"}

	envoy, err := setupStack(base)
	if err != nil {
		t.Fatal(err)
	}
	proj, err := envoy.Project()
	if err != nil {
		t.Fatal(err)
	}
	gw := proj.Services["api-gw"]
	if gw.ContainerName != "supabase-envoy" {
		t.Fatalf("envoy container = %q", gw.ContainerName)
	}
	if proj.Networks["default"].Name != "supabase_default" {
		t.Fatalf("engine network = %q", proj.Networks["default"].Name)
	}
	if aliases := gw.Networks["default"].Aliases; !sameSet(aliases, []string{"envoy", "kong"}) {
		t.Fatalf("gateway aliases = %v", aliases)
	}
	if !published(gw, "8000", 8000) || published(gw, "8443", 8443) {
		t.Fatalf("envoy ports = %+v", gw.Ports)
	}
	assertFileMount(t, proj.Services["db"], "/_supabase.sql", "/docker-entrypoint-initdb.d/migrations/97-_supabase.sql")
	for _, vol := range proj.Services["db"].Volumes {
		if vol.Target == "/docker-entrypoint-initdb.d/migrations" || vol.Target == "/docker-entrypoint-initdb.d/init-scripts" {
			t.Fatalf("init directory is hidden by %s", vol.Target)
		}
	}
	if proj.Services["functions"].MemLimit == 0 {
		t.Fatal("functions memory limit was not set")
	}
	if proj.Services["realtime"].ContainerName != "realtime-dev.supabase-realtime" {
		t.Fatalf("realtime container = %q", proj.Services["realtime"].ContainerName)
	}

	caddyCfg := base
	caddyCfg.gateway = "caddy"
	caddy, err := setupStack(caddyCfg)
	if err != nil {
		t.Fatal(err)
	}
	caddyProj, err := caddy.Project()
	if err != nil {
		t.Fatal(err)
	}
	if len(caddyProj.Services["api-gw"].Ports) != 0 {
		t.Fatalf("caddy mode published the gateway: %+v", caddyProj.Services["api-gw"].Ports)
	}
	if !published(caddyProj.Services["caddy"], "443", 443) {
		t.Fatalf("caddy ports = %+v", caddyProj.Services["caddy"].Ports)
	}

	kongCfg := base
	kongCfg.gateway = "kong"
	kong, err := setupStack(kongCfg)
	if err != nil {
		t.Fatal(err)
	}
	kongProj, err := kong.Project()
	if err != nil {
		t.Fatal(err)
	}
	if kongProj.Services["api-gw"].ContainerName != "supabase-kong" {
		t.Fatalf("kong container = %q", kongProj.Services["api-gw"].ContainerName)
	}
	if !published(kongProj.Services["api-gw"], "8443", 8443) {
		t.Fatalf("kong ports = %+v", kongProj.Services["api-gw"].Ports)
	}

	s3Cfg := base
	s3Cfg.storage = "s3"
	s3, err := setupStack(s3Cfg)
	if err != nil {
		t.Fatal(err)
	}
	s3Proj, err := s3.Project()
	if err != nil {
		t.Fatal(err)
	}
	backend := s3Proj.Services["storage"].Environment["STORAGE_BACKEND"]
	if backend == nil || *backend != "s3" {
		t.Fatalf("storage backend = %v", backend)
	}
	if _, ok := s3Proj.Services["storage"].DependsOn["minio-createbucket"]; !ok {
		t.Fatal("s3 storage does not wait for the bucket job")
	}
}

func published(svc composetypes.ServiceConfig, host string, container uint32) bool {
	for _, port := range svc.Ports {
		if port.Published == host && port.Target == container {
			return true
		}
	}
	return false
}

func assertFileMount(t *testing.T, svc composetypes.ServiceConfig, sourceSuffix, target string) {
	t.Helper()
	for _, vol := range svc.Volumes {
		if vol.Target == target && strings.HasSuffix(vol.Source, sourceSuffix) {
			return
		}
	}
	t.Fatalf("missing mount %s -> %s in %+v", sourceSuffix, target, svc.Volumes)
}

func sameSet(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	seen := map[string]int{}
	for _, v := range got {
		seen[v]++
	}
	for _, v := range want {
		seen[v]--
		if seen[v] < 0 {
			return false
		}
	}
	return true
}
