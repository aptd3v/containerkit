package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/moby/moby/api/types/registry"
	"github.com/stretchr/testify/require"
)

func TestAuthToBase64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   Auth
	}{
		{name: "empty", in: Auth{}},
		{name: "docker hub", in: Auth{Username: "u", Password: "p", ServerAddress: "https://index.docker.io/v1/"}},
		{name: "special password", in: Auth{Username: "u", Password: "p+/=?&", ServerAddress: "localhost:5000"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			enc, err := AuthToBase64(tt.in)
			require.NoError(t, err)
			raw, err := base64.URLEncoding.DecodeString(enc)
			require.NoError(t, err)

			var keys map[string]string
			require.NoError(t, json.Unmarshal(raw, &keys))
			require.Equal(t, tt.in.Username, keys["username"])
			require.Equal(t, tt.in.Password, keys["password"])
			require.Equal(t, tt.in.ServerAddress, keys["serveraddress"])

			var got Auth
			require.NoError(t, json.Unmarshal(raw, &got))
			require.Equal(t, tt.in, got)
		})
	}
}

func TestRegistryAuthConfigOptions(t *testing.T) {
	t.Parallel()
	creds := Auth{Username: "u", Password: "p", ServerAddress: "docker.io"}
	wantAuth, err := AuthToBase64(creds)
	require.NoError(t, err)

	cfg := &registry.AuthConfig{}
	require.NoError(t, WithUsername("user")(cfg))
	require.NoError(t, WithPassword("pass")(cfg))
	require.NoError(t, WithAuth(creds)(cfg))
	require.NoError(t, WithEmail("")(cfg))
	require.Error(t, WithEmail("e@x")(cfg))
	require.NoError(t, WithServerAddress("reg")(cfg))
	require.NoError(t, WithIdentityToken("id")(cfg))
	require.NoError(t, WithRegistryToken("tok")(cfg))

	require.Equal(t, "user", cfg.Username)
	require.Equal(t, "pass", cfg.Password)
	require.Equal(t, wantAuth, cfg.Auth)
	require.Equal(t, "reg", cfg.ServerAddress)
	require.Equal(t, "id", cfg.IdentityToken)
	require.Equal(t, "tok", cfg.RegistryToken)
}

func ExampleAuthToBase64() {
	s, err := AuthToBase64(Auth{Username: "u", Password: "p", ServerAddress: "docker.io"})
	if err != nil {
		panic(err)
	}
	raw, _ := base64.URLEncoding.DecodeString(s)
	fmt.Println(string(raw))
	// Output: {"username":"u","password":"p","serveraddress":"docker.io"}
}
