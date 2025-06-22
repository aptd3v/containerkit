package auth

import (
	"encoding/base64"
	"encoding/json"

	"github.com/docker/docker/api/types/registry"
)

// Auth represents registry authentication credentials
type Auth struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	ServerAddress string `json:"serveraddress"`
}

// AuthToBase64 converts auth credentials to base64 encoded auth string
func AuthToBase64(auth Auth) (string, error) {
	jsonAuth, err := json.Marshal(auth)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(jsonAuth), nil
}

// SetAuthConfigOption is a function that sets a parameter for the registry auth config.
type SetRegistryAuthConfigOption func(*registry.AuthConfig) error

// WithUsername is the username for the registry.
func WithUsername(username string) SetRegistryAuthConfigOption {
	return func(o *registry.AuthConfig) error {
		o.Username = username
		return nil
	}
}

// WithPassword is the password for the registry.
func WithPassword(password string) SetRegistryAuthConfigOption {
	return func(o *registry.AuthConfig) error {
		o.Password = password
		return nil
	}
}

// WithAuth is the base64 encoded auth string.
func WithAuth(creds Auth) SetRegistryAuthConfigOption {
	return func(o *registry.AuthConfig) error {
		auth, err := AuthToBase64(creds)
		if err != nil {
			return err
		}
		o.Auth = auth
		return nil
	}
}

// WithEmail is an optional setter associated with the username.
//
// Deprecated: will be removed in a later version of docker.
func WithEmail(email string) SetRegistryAuthConfigOption {
	return func(o *registry.AuthConfig) error {
		o.Email = email
		return nil
	}
}

// WithServerAddress is the address of the registry.
func WithServerAddress(serverAddress string) SetRegistryAuthConfigOption {
	return func(o *registry.AuthConfig) error {
		o.ServerAddress = serverAddress
		return nil
	}
}

// WithIdentityToken is used to authenticate the user and get an access token for the registry.
func WithIdentityToken(identityToken string) SetRegistryAuthConfigOption {
	return func(o *registry.AuthConfig) error {
		o.IdentityToken = identityToken
		return nil
	}
}

// WithRegistryToken is a bearer token to be sent to a registry
func WithRegistryToken(registryToken string) SetRegistryAuthConfigOption {
	return func(o *registry.AuthConfig) error {
		o.RegistryToken = registryToken
		return nil
	}
}
