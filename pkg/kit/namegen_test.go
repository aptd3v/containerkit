package kit

import "testing"

func TestSanitizeName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want string
	}{
		{in: "Nginx", want: "nginx"},
		{in: "my.app", want: "my-app"},
		{in: "...dots...", want: "dots"},
		{in: "123abc", want: "s123abc"},
		{in: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			if got := sanitizeName(tt.in); got != tt.want {
				t.Fatalf("sanitizeName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestImageBasename(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want string
	}{
		{in: "nginx:latest", want: "nginx"},
		{in: "ghcr.io/org/app:1", want: "app"},
		{in: "redis@sha256:dead", want: "redis"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			if got := imageBasename(tt.in); got != tt.want {
				t.Fatalf("imageBasename(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestUniqueName(t *testing.T) {
	t.Parallel()
	taken := map[string]bool{"nginx": true, "nginx-2": true}
	got := uniqueName("nginx", func(n string) bool { return taken[n] })
	if got != "nginx-3" {
		t.Fatalf("uniqueName = %q", got)
	}
}
