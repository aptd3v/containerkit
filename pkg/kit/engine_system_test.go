package kit

import (
	"context"
	"testing"

	"github.com/aptd3v/containerkit/pkg/df"
)

func TestEngineSystemMethodsAndSetters(t *testing.T) {
	t.Parallel()
	api := &fakeEngine{}
	ctx := context.Background()

	ping, err := Ping(ctx, api, NegotiateAPIVersion(), ForceNegotiate(true))
	if err != nil {
		t.Fatal(err)
	}
	if ping.APIVersion != "1.56" {
		t.Fatalf("ping = %+v", ping)
	}
	if !api.ping.NegotiateAPIVersion || !api.ping.ForceNegotiate {
		t.Fatalf("ping opts = %+v", api.ping)
	}

	if _, err := Info(ctx, api); err != nil {
		t.Fatal(err)
	}

	ver, err := EngineVersion(ctx, api)
	if err != nil {
		t.Fatal(err)
	}
	if ver.Version != "29.0.0" || ver.APIVersion != "1.56" {
		t.Fatalf("version = %+v", ver)
	}

	gotVer, err := ClientVersion(api)
	if err != nil {
		t.Fatal(err)
	}
	if gotVer != "1.56" {
		t.Fatalf("client version = %q", gotVer)
	}
	host, err := DaemonHost(api)
	if err != nil {
		t.Fatal(err)
	}
	if host != "unix:///var/run/docker.sock" {
		t.Fatalf("daemon host = %q", host)
	}

	if _, err := DiskUsage(ctx, api, Verbose(true), Usage(df.Container), Images(true), BuildCache(true), Volumes(true)); err != nil {
		t.Fatal(err)
	}
	if !api.diskUsage.Verbose || !api.diskUsage.Containers || !api.diskUsage.Images || !api.diskUsage.BuildCache || !api.diskUsage.Volumes {
		t.Fatalf("disk usage flags = %+v", api.diskUsage)
	}

	api.diskUsage = DiskUsageOptions{}
	if _, err := DiskUsage(ctx, api, Usage(df.Container), Usage(df.Image), Usage(df.Volume), Usage(df.BuildCache)); err != nil {
		t.Fatal(err)
	}
	if !api.diskUsage.Containers || !api.diskUsage.Images || !api.diskUsage.Volumes || !api.diskUsage.BuildCache {
		t.Fatalf("disk usage types = %+v", api.diskUsage)
	}

	if _, err := Login(ctx, api, Username("alice"), Password("secret"), ServerAddress("registry.example"), IdentityToken("idtok"), RegistryToken("regtok")); err != nil {
		t.Fatal(err)
	}
	if api.login.Username != "alice" || api.login.Password != "secret" || api.login.ServerAddress != "registry.example" || api.login.IdentityToken != "idtok" || api.login.RegistryToken != "regtok" {
		t.Fatalf("login = %+v", api.login)
	}

	_ = ListenEvents(ctx, api, Since("1h"), Until("now"), Filter("type", "container"))
	if api.events.Since != "1h" || api.events.Until != "now" {
		t.Fatalf("events = %+v", api.events)
	}
	if !api.events.Filters["type"]["container"] {
		t.Fatalf("events filters = %+v", api.events.Filters)
	}

	grouped := &fakeEngine{}
	if _, err := Ping(ctx, grouped, GroupedSet([]any{NegotiateAPIVersion(), ForceNegotiate(true)})); err != nil {
		t.Fatal(err)
	}
	if !grouped.ping.NegotiateAPIVersion || !grouped.ping.ForceNegotiate {
		t.Fatalf("grouped ping = %+v", grouped.ping)
	}
	if _, err := DiskUsage(ctx, grouped, GroupedSet([]any{Usage(df.Image), Verbose(true)})); err != nil {
		t.Fatal(err)
	}
	if !grouped.diskUsage.Images || !grouped.diskUsage.Verbose {
		t.Fatalf("grouped disk usage = %+v", grouped.diskUsage)
	}
	if _, err := Login(ctx, grouped, GroupedSet([]any{Username("bob"), Password("pw")})); err != nil {
		t.Fatal(err)
	}
	if grouped.login.Username != "bob" || grouped.login.Password != "pw" {
		t.Fatalf("grouped login = %+v", grouped.login)
	}
}

func TestEngineSystemNilClientAndInvalidSetters(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	api := &fakeEngine{}
	tests := []struct {
		name string
		call func() error
	}{
		{"ping nil engine", func() error { _, err := Ping(ctx, nil); return err }},
		{"info nil engine", func() error { _, err := Info(ctx, nil); return err }},
		{"version nil engine", func() error { _, err := EngineVersion(ctx, nil); return err }},
		{"client version nil engine", func() error { _, err := ClientVersion(nil); return err }},
		{"daemon host nil engine", func() error { _, err := DaemonHost(nil); return err }},
		{"disk usage nil engine", func() error { _, err := DiskUsage(ctx, nil); return err }},
		{"login nil engine", func() error { _, err := Login(ctx, nil); return err }},
		{"nil ping option", func() error { _, err := Ping(ctx, api, nil); return err }},
		{"nil disk usage option", func() error { _, err := DiskUsage(ctx, api, nil); return err }},
		{"nil login option", func() error { _, err := Login(ctx, api, nil); return err }},
		{"empty username", func() error { _, err := Login(ctx, api, Username("")); return err }},
		{"empty password", func() error { _, err := Login(ctx, api, Password("")); return err }},
		{"empty server address", func() error { _, err := Login(ctx, api, ServerAddress("")); return err }},
		{"empty identity token", func() error { _, err := Login(ctx, api, IdentityToken("")); return err }},
		{"empty registry token", func() error { _, err := Login(ctx, api, RegistryToken("")); return err }},
		{"invalid usage type", func() error { _, err := DiskUsage(ctx, api, Usage(df.Object("nope"))); return err }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err == nil {
				t.Fatal("expected error")
			}
		})
	}
	if ev := ListenEvents(ctx, nil); ev.Messages != nil || ev.Err != nil {
		t.Fatalf("nil engine events = %+v", ev)
	}
}
