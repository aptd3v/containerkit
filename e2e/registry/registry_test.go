//go:build e2e

package registry_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/aptd3v/containerkit/pkg/kit"
)

const (
	seedImage = "alpine:3.21"
	repoName  = "containerkit/e2e-auth"
	tagName   = "latest"
)

func TestEngineRegistryAuthentication(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	t.Cleanup(cancel)

	engineHost, registryHost, user, pass := registryEnv(t)
	api, err := kit.NewEngine(kit.Host(engineHost))
	if err != nil {
		t.Skipf("NewEngine(%s): %v", engineHost, err)
	}
	t.Cleanup(func() { _ = api.Close() })
	if _, err := kit.Ping(ctx, api); err != nil {
		t.Skipf("engine not reachable at %s: %v", engineHost, err)
	}

	ref := registryHost + "/" + repoName + ":" + tagName
	auth := kit.RegistryAuth(kit.EncodeRegistryAuth(user, pass, registryHost))
	log := kit.NewLogger(kit.OutputStream(io.Discard), kit.Quiet(true))

	if _, err := kit.Login(ctx, api, kit.Username(user), kit.Password("wrong-password"), kit.ServerAddress(registryHost)); err == nil {
		t.Fatal("Login with wrong password succeeded")
	}

	login, err := kit.Login(ctx, api, kit.Username(user), kit.Password(pass), kit.ServerAddress(registryHost))
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if login.Auth.Status == "" && login.Auth.IdentityToken == "" {
		t.Fatalf("Login returned empty auth: %+v", login.Auth)
	}
	t.Logf("login status=%q identity=%t", login.Auth.Status, login.Auth.IdentityToken != "")

	if _, err := kit.PullImage(ctx, api, seedImage, kit.Log(log)); err != nil {
		t.Fatalf("PullImage %s: %v", seedImage, err)
	}
	if _, err := kit.TagImage(ctx, api, kit.Source(seedImage), kit.Target(ref)); err != nil {
		t.Fatalf("TagImage: %v", err)
	}

	if _, err := kit.PushImage(ctx, api, ref, kit.Log(log)); err == nil {
		t.Fatal("PushImage without credentials succeeded")
	}

	if _, err := kit.PushImage(ctx, api, ref, auth, kit.Log(log)); err != nil {
		t.Fatalf("PushImage with credentials: %v", err)
	}

	if _, err := kit.RemoveImage(ctx, api, ref, kit.Force(true)); err != nil {
		t.Fatalf("RemoveImage %s: %v", ref, err)
	}

	if _, err := kit.PullImage(ctx, api, ref, kit.Log(log)); err == nil {
		t.Fatal("PullImage without credentials succeeded")
	}
	if _, err := kit.PullImage(ctx, api, ref, auth, kit.Log(log)); err != nil {
		t.Fatalf("PullImage with credentials: %v", err)
	}
	if _, err := kit.InspectImage(ctx, api, ref); err != nil {
		t.Fatalf("InspectImage after authenticated pull: %v", err)
	}

	if login.Auth.IdentityToken != "" {
		tok := kit.RegistryAuth(kit.EncodeRegistryIdentity(login.Auth.IdentityToken, registryHost))
		if _, err := kit.PushImage(ctx, api, ref, tok, kit.Log(log)); err != nil {
			t.Fatalf("PushImage with identity token: %v", err)
		}
	}

	t.Cleanup(func() {
		_, _ = kit.RemoveImage(context.Background(), api, ref, kit.Force(true))
	})
}

func registryEnv(t *testing.T) (engineHost, registryHost, user, pass string) {
	t.Helper()
	loadDotEnv(t)
	engineHost = os.Getenv("CONTAINERKIT_REGISTRY_ENGINE")
	registryHost = os.Getenv("CONTAINERKIT_REGISTRY_HOST")
	user = os.Getenv("CONTAINERKIT_REGISTRY_USER")
	pass = os.Getenv("CONTAINERKIT_REGISTRY_PASSWORD")
	if engineHost == "" {
		engineHost = "tcp://127.0.0.1:23755"
	}
	if registryHost == "" {
		registryHost = "127.0.0.1:5000"
	}
	if user == "" {
		user = "containerkit"
	}
	if pass == "" {
		pass = "containerkit-e2e"
	}
	return engineHost, registryHost, user, pass
}

func loadDotEnv(t *testing.T) {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return
	}
	envf := filepath.Join(filepath.Dir(file), ".run", "env")
	b, err := os.ReadFile(envf)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		if os.Getenv(k) == "" {
			t.Setenv(k, v)
		}
	}
}
