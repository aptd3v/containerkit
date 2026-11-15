package main

import "testing"

func TestBagTSName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
	}{
		{"UpOption", "UpOptions"},
		{"ComposerClientOption", "ConnectOptions"},
		{"EnginePullOption", "PullImageOptions"},
		{"EngineImageListOption", "ListImagesOptions"},
		{"TLSOption", "TLSOptions"},
		{"ClientOption", "ClientOptions"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			if got := bagTSName(tc.in); got != tc.want {
				t.Fatalf("bagTSName(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
