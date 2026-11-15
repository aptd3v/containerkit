//go:build e2e

package model_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/aptd3v/containerkit/pkg/kit"
)

const (
	projectName = "containerkit-e2e-embed"
	serviceName = "client"
	modelName   = "embedding"
	// Official Docker Model Runner artifact for Google EmbeddingGemma.
	// Output dimensionality is 768, matching Gemini Embedding 2's recommended size.
	embeddingImage = "ai/embeddinggemma:300M-Q8_0"
	embeddingDim   = 768
	clientImage    = "alpine:3.21"
	embedText      = "Gemma maps this sentence into a dense vector."
	dmrHostPort    = "12434"
)

func TestComposeEmbeddingGemma(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	t.Cleanup(cancel)

	api, err := kit.NewEngine()
	if err != nil {
		t.Skipf("NewEngine: %v", err)
	}
	t.Cleanup(func() { _ = api.Close() })
	if _, err := kit.Ping(ctx, api); err != nil {
		t.Skipf("engine not reachable: %v", err)
	}

	stack := embeddingStack()
	if err := stack.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	yml, err := stack.ExportYAML()
	if err != nil {
		t.Fatalf("ExportYAML: %v", err)
	}
	text := string(yml)
	for _, want := range []string{
		embeddingImage,
		"--embeddings",
		"context_size: 2048",
		"endpoint_var: EMBEDDING_URL",
		"model_var: EMBEDDING_MODEL",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("export missing %q in:\n%s", want, text)
		}
	}

	if err := modelRunnerReady(ctx); err != nil {
		t.Skipf("docker model runner unavailable: %v", err)
	}
	composer, err := kit.NewComposer()
	if err != nil {
		t.Fatalf("NewComposer: %v", err)
	}

	t.Cleanup(func() {
		downCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Minute)
		defer cancel()
		if err := stack.Down(downCtx, composer, kit.RemoveOrphans(true)); err != nil {
			t.Logf("Down: %v", err)
		}
	})

	if err := stack.Up(ctx, composer, kit.Wait(true)); err != nil {
		if modelUnavailable(err, nil) {
			t.Skipf("embedding model unavailable: %v", err)
		}
		t.Fatalf("Up: %v", err)
	}

	endpoint, model, err := injectedModelEnv(ctx, api)
	if err != nil {
		t.Fatal(err)
	}
	if endpoint == "" {
		t.Fatal("EMBEDDING_URL was not injected")
	}
	if model == "" {
		t.Fatal("EMBEDDING_MODEL was not injected")
	}
	if !strings.Contains(strings.ToLower(model), "embeddinggemma") {
		t.Fatalf("EMBEDDING_MODEL = %q, want embeddinggemma", model)
	}

	vec, err := embed(ctx, rewriteModelURL(endpoint), model, embedText)
	if err != nil {
		t.Fatalf("embeddings: %v", err)
	}
	if len(vec) != embeddingDim {
		t.Fatalf("embedding dim = %d, want %d", len(vec), embeddingDim)
	}
	var nonzero bool
	for _, v := range vec {
		if v != 0 {
			nonzero = true
			break
		}
	}
	if !nonzero {
		t.Fatal("embedding vector is all zeros")
	}
}

func embeddingStack() *kit.Compose {
	return kit.NewCompose(projectName,
		kit.AddModel(
			kit.Name(modelName),
			kit.Image(embeddingImage),
			kit.ContextSize(2048),
			kit.RuntimeFlag("--embeddings"),
		),
	).WithService(kit.NewService(
		kit.NewContainer(
			kit.Name(serviceName),
			kit.Image(clientImage),
			kit.Command("sleep", "86400"),
		),
		kit.Model(
			kit.Name(modelName),
			kit.EndpointVar("EMBEDDING_URL"),
			kit.ModelVar("EMBEDDING_MODEL"),
		),
	))
}

func injectedModelEnv(ctx context.Context, api kit.Engine) (endpoint, model string, err error) {
	listed, err := kit.ListContainers(ctx, api,
		kit.All(true),
		kit.Filter("label", "com.docker.compose.project="+projectName),
		kit.Filter("label", "com.docker.compose.service="+serviceName),
	)
	if err != nil {
		return "", "", fmt.Errorf("list containers: %w", err)
	}
	if len(listed.Items) == 0 {
		return "", "", fmt.Errorf("service %q container not found", serviceName)
	}
	ins, err := new(kit.Container).Inspect(ctx, api, listed.Items[0].ID)
	if err != nil {
		return "", "", fmt.Errorf("inspect: %w", err)
	}
	if ins.Container.Config == nil {
		return "", "", fmt.Errorf("inspect: empty container config")
	}
	return envVal(ins.Container.Config.Env, "EMBEDDING_URL"), envVal(ins.Container.Config.Env, "EMBEDDING_MODEL"), nil
}

func embed(ctx context.Context, base, model, input string) ([]float64, error) {
	body, err := json.Marshal(map[string]any{
		"model":           model,
		"input":           input,
		"encoding_format": "float",
	})
	if err != nil {
		return nil, err
	}
	endpoint := strings.TrimRight(base, "/") + "/embeddings"
	var last error
	for attempt := 0; attempt < 12; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(5 * time.Second):
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			last = err
			continue
		}
		raw, readErr := io.ReadAll(res.Body)
		_ = res.Body.Close()
		if readErr != nil {
			last = readErr
			continue
		}
		if res.StatusCode >= 500 || res.StatusCode == http.StatusNotFound {
			last = fmt.Errorf("%s: %s", res.Status, bytes.TrimSpace(raw))
			continue
		}
		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("%s: %s", res.Status, bytes.TrimSpace(raw))
		}
		var out struct {
			Data []struct {
				Embedding []float64 `json:"embedding"`
			} `json:"data"`
		}
		if err := json.Unmarshal(raw, &out); err != nil {
			return nil, fmt.Errorf("decode embeddings: %w: %s", err, bytes.TrimSpace(raw))
		}
		if len(out.Data) == 0 || len(out.Data[0].Embedding) == 0 {
			return nil, fmt.Errorf("empty embedding payload: %s", bytes.TrimSpace(raw))
		}
		return out.Data[0].Embedding, nil
	}
	if last == nil {
		last = fmt.Errorf("embeddings endpoint %s did not become ready", endpoint)
	}
	return nil, last
}

func rewriteModelURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return strings.TrimRight(raw, "/")
	}
	switch strings.ToLower(u.Hostname()) {
	case "model-runner.docker.internal", "host.docker.internal":
		port := u.Port()
		if port == "" {
			port = dmrHostPort
		}
		u.Host = net.JoinHostPort("127.0.0.1", port)
	}
	return strings.TrimRight(u.String(), "/")
}

func envVal(env []string, key string) string {
	prefix := key + "="
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			return strings.TrimPrefix(e, prefix)
		}
	}
	return ""
}

func modelRunnerReady(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "model", "status", "--json")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, bytes.TrimSpace(out))
	}
	var st struct {
		Running bool `json:"running"`
	}
	if err := json.Unmarshal(out, &st); err != nil {
		return fmt.Errorf("status json: %w", err)
	}
	if !st.Running {
		return fmt.Errorf("docker model runner is not running")
	}
	return nil
}

func modelUnavailable(err error, out []byte) bool {
	msg := strings.ToLower(err.Error() + " " + string(out))
	for _, n := range []string{
		"model plugin",
		"model runner",
		"docker-model",
		"llama.cpp",
		"not installed",
		"denied",
		"403",
		"connection refused",
	} {
		if strings.Contains(msg, n) {
			return true
		}
	}
	return false
}
