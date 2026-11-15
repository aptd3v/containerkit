package cap

import "testing"

func TestAllKnownCapabilities(t *testing.T) {
	t.Parallel()
	seen := map[Capability]struct{}{}
	for _, c := range All() {
		if c == "" {
			t.Fatal("empty capability")
		}
		if _, ok := seen[c]; ok {
			t.Fatalf("duplicate %s", c)
		}
		seen[c] = struct{}{}
	}
	if _, ok := seen[NET_ADMIN]; !ok {
		t.Fatal("missing NET_ADMIN")
	}
	if _, ok := seen[ALL]; !ok {
		t.Fatal("missing ALL")
	}
	if len(seen) != 42 {
		t.Fatalf("got %d capabilities, want 42 (ALL + 41 kernel caps)", len(seen))
	}
}
