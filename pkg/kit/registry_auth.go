package kit

import (
	"encoding/base64"
	"encoding/json"
)

type registryAuthJSON struct {
	Username      string `json:"username,omitempty"`
	Password      string `json:"password,omitempty"`
	ServerAddress string `json:"serveraddress,omitempty"`
	IdentityToken string `json:"identitytoken,omitempty"`
	RegistryToken string `json:"registrytoken,omitempty"`
}

func encodeRegistryAuth(cfg registryAuthJSON) string {
	buf, err := json.Marshal(cfg)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(buf)
}

// EncodeRegistryAuth returns the X-Registry-Auth value for [RegistryAuth]
// (base64url JSON username/password/serveraddress).
func EncodeRegistryAuth(username, password, serverAddress string) string {
	return encodeRegistryAuth(registryAuthJSON{
		Username:      username,
		Password:      password,
		ServerAddress: serverAddress,
	})
}

// EncodeRegistryIdentity returns the X-Registry-Auth value for [RegistryAuth]
// from a login identity token.
func EncodeRegistryIdentity(identityToken, serverAddress string) string {
	return encodeRegistryAuth(registryAuthJSON{
		IdentityToken: identityToken,
		ServerAddress: serverAddress,
	})
}
