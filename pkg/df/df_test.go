package df

import "testing"

func TestAllDiskUsageObjects(t *testing.T) {
	t.Parallel()
	want := []Object{Container, Image, Volume, BuildCache}
	got := All()
	if len(got) != len(want) {
		t.Fatalf("All() = %v, want %v", got, want)
	}
	seen := map[Object]struct{}{}
	for i, o := range got {
		if o != want[i] {
			t.Fatalf("All()[%d] = %q, want %q", i, o, want[i])
		}
		if o == "" {
			t.Fatal("empty object")
		}
		if _, ok := seen[o]; ok {
			t.Fatalf("duplicate %q", o)
		}
		seen[o] = struct{}{}
	}
	if Container != "container" || Image != "image" || Volume != "volume" || BuildCache != "build-cache" {
		t.Fatalf("values = %q %q %q %q", Container, Image, Volume, BuildCache)
	}
}
