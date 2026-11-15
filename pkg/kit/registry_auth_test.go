package kit

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestEncodeRegistryAuth(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		encode func() string
		want   registryAuthJSON
	}{
		{
			name:   "username password",
			encode: func() string { return EncodeRegistryAuth("alice", "secret", "127.0.0.1:5000") },
			want:   registryAuthJSON{Username: "alice", Password: "secret", ServerAddress: "127.0.0.1:5000"},
		},
		{
			name:   "identity token",
			encode: func() string { return EncodeRegistryIdentity("idtok", "registry.example") },
			want:   registryAuthJSON{IdentityToken: "idtok", ServerAddress: "registry.example"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.encode()
			if got == "" {
				t.Fatal("empty encoding")
			}
			raw, err := base64.URLEncoding.DecodeString(got)
			if err != nil {
				t.Fatal(err)
			}
			var cfg registryAuthJSON
			if err := json.Unmarshal(raw, &cfg); err != nil {
				t.Fatal(err)
			}
			if cfg != tt.want {
				t.Fatalf("decoded %+v, want %+v", cfg, tt.want)
			}
		})
	}
}
