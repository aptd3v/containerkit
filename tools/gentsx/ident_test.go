package main

import "testing"

func TestJSIdent(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
	}{
		{"Service", "service"},
		{"GPU", "gpu"},
		{"SSH", "ssh"},
		{"IPAMOpt", "ipamOpt"},
		{"HostPort", "hostPort"},
		{"fileConfig", "fileConfig"},
		{"DependsOn", "dependsOn"},
		{"WatchFiles", "watchFiles"},
		{"CPUs", "cpus"},
		{"IPv4", "ipv4"},
		{"IPv6", "ipv6"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			if got := jsIdent(tc.in); got != tc.want {
				t.Fatalf("jsIdent(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestDropComposeDuplicate(t *testing.T) {
	t.Parallel()
	kept := verb{tsName: "composeUp", goName: "Up", recv: "Client"}
	extra := verb{tsName: "composeUp", goName: "Up", recv: "Compose"}
	if !dropComposeDuplicate(kept, extra) {
		t.Fatal("Compose.Up should be dropped when Client.Up already covers it")
	}
	other := verb{tsName: "composeRun", goName: "Run", recv: "Compose"}
	if dropComposeDuplicate(kept, other) {
		t.Fatal("a different Go method should be kept")
	}
}

func TestComposeCommandName(t *testing.T) {
	t.Parallel()
	if got := composeCommandName("up", true); got != "composeUp" {
		t.Fatalf("up = %q", got)
	}
	if got := composeCommandName("composeExec", true); got != "composeExec" {
		t.Fatalf("composeExec = %q", got)
	}
	if got := composeCommandName("pullImage", false); got != "pullImage" {
		t.Fatalf("pullImage = %q", got)
	}
	if got := composeCommandName("publishOCI", true); got != "composePublishOCI" {
		t.Fatalf("publishOCI = %q", got)
	}
}

func TestJSXPropName(t *testing.T) {
	t.Parallel()
	if got := jsxPropName("key"); got != "name" {
		t.Fatalf("key = %q", got)
	}
	if got := jsxPropName("type"); got != "kind" {
		t.Fatalf("type = %q", got)
	}
}
