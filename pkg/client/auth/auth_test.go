package auth

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/docker/docker/api/types/registry"
	"github.com/stretchr/testify/require"
)

func TestAuthToBase64AndSetters(t *testing.T) {
	a := Auth{Username: "u", Password: "p", ServerAddress: "docker.io"}
	enc, err := AuthToBase64(a)
	require.NoError(t, err)
	raw, err := base64.URLEncoding.DecodeString(enc)
	require.NoError(t, err)
	var got Auth
	require.NoError(t, json.Unmarshal(raw, &got))
	require.Equal(t, a, got)

	cfg := &registry.AuthConfig{}
	require.NoError(t, WithUsername("user")(cfg))
	require.NoError(t, WithPassword("pass")(cfg))
	require.NoError(t, WithAuth(a)(cfg))
	require.NoError(t, WithEmail("e@x")(cfg))
	require.NoError(t, WithServerAddress("reg")(cfg))
	require.NoError(t, WithIdentityToken("id")(cfg))
	require.NoError(t, WithRegistryToken("tok")(cfg))
	require.Equal(t, "user", cfg.Username)
	require.Equal(t, "pass", cfg.Password)
	require.Equal(t, "e@x", cfg.Email)
	require.Equal(t, "reg", cfg.ServerAddress)
	require.Equal(t, "id", cfg.IdentityToken)
	require.Equal(t, "tok", cfg.RegistryToken)
	require.NotEmpty(t, cfg.Auth)
}
