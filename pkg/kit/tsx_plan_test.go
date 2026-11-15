package kit

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestImportPlan(t *testing.T) {
	t.Parallel()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller")
	}
	path := filepath.Join(filepath.Dir(file), "..", "..", "tsx", "testdata", "plan.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	p := ImportPlan(data)
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	c := p.Compose()
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	if c.Name() != "app" {
		t.Fatalf("name = %q", c.Name())
	}
	svcs := c.Services()
	if len(svcs) != 1 || svcs[0].Name() != "api" || svcs[0].Image() != "alpine:latest" {
		t.Fatalf("services = %+v", svcs)
	}
}

func TestDecodeConnectAndUpOptions(t *testing.T) {
	t.Parallel()
	connect, err := decodeConnectOptions(map[string]any{
		"fromEnv":   true,
		"host":      "unix:///var/run/docker.sock",
		"timeout":   "10s",
		"userAgent": "test",
		"tls": map[string]any{
			"ca":   "/ca.pem",
			"cert": "/cert.pem",
			"key":  "/key.pem",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(connect) < 4 {
		t.Fatalf("connect opts = %d", len(connect))
	}
	up, err := decodeUpOptions(map[string]any{
		"detach":  true,
		"timeout": "30s",
		"quiet":   true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(up) != 3 {
		t.Fatalf("up opts = %d", len(up))
	}
	d, err := asDuration("10s")
	if err != nil || d != 10*time.Second {
		t.Fatalf("duration = %v, %v", d, err)
	}
}

func TestDecodeVocabOptions(t *testing.T) {
	t.Parallel()
	kill, err := decodeKillOptions(map[string]any{"stopSignal": "SIGTERM"})
	if err != nil || len(kill) != 1 {
		t.Fatalf("stopSignal: %v (%d)", err, len(kill))
	}
	published, err := decodePublishOptions(map[string]any{"oci": "1.1"})
	if err != nil || len(published) != 1 {
		t.Fatalf("oci: %v (%d)", err, len(published))
	}
	updated, err := decodeUpdateOptions(map[string]any{"order": "start-first"})
	if err != nil || len(updated) != 1 {
		t.Fatalf("order: %v (%d)", err, len(updated))
	}
	volume, err := decodeVolumeOptions(map[string]any{"seLinux": "z"})
	if err != nil || len(volume) != 1 {
		t.Fatalf("seLinux: %v (%d)", err, len(volume))
	}
	inspected, err := decodeEngineImageInspectOptions(map[string]any{
		"imagePlatform": map[string]any{"architecture": "amd64", "os": "linux"},
	})
	if err != nil || len(inspected) != 1 {
		t.Fatalf("imagePlatform: %v (%d)", err, len(inspected))
	}
	if _, err := decodeEngineImageInspectOptions(map[string]any{"imagePlatform": "linux/amd64"}); err == nil {
		t.Fatal("imagePlatform string was accepted")
	}
}

func TestDockerfileTSXPlan(t *testing.T) {
	t.Parallel()
	yml := []byte("name: app\nservices:\n  api:\n    build:\n      dockerfile_inline: \"FROM alpine:latest\\nRUN echo ok\\n\"\n")
	c := ImportYAML(yml)
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	out, err := c.ExportYAML()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "FROM alpine:latest") || !strings.Contains(string(out), "RUN echo ok") {
		t.Fatalf("export = %s", out)
	}

	plan := ImportPlan([]byte(`{"connect":{"fromEnv":true},"actions":[{"op":"buildImage","options":{"image":"app:dev","parseDockerfile":"FROM alpine:latest\nCMD [\"true\"]\n"}}]}`))
	if err := plan.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestDrainedStreamPlansReturnOK(t *testing.T) {
	t.Parallel()
	cli := NewClient(&fakeEngine{}, nil)
	cases := []planAction{
		{Op: "pullImage", Ref: "alpine:latest"},
		{Op: "pushImage", Ref: "alpine:latest"},
		{Op: "buildImage", Options: map[string]any{"parseDockerfile": "FROM alpine:latest\n"}},
	}
	for _, act := range cases {
		out, err := execKnownAction(t.Context(), cli, nil, act)
		if err != nil {
			t.Fatalf("%s: %v", act.Op, err)
		}
		if string(out) != `{"ok":true}` {
			t.Fatalf("%s = %s", act.Op, out)
		}
	}
}

func TestExecDispatchesKitVerbs(t *testing.T) {
	t.Parallel()
	compose := NewCompose("app")
	_, err := execKnownAction(t.Context(), nil, compose, planAction{Op: "not-a-verb"})
	if err == nil || !strings.Contains(err.Error(), "unknown op") {
		t.Fatalf("unknown op: %v", err)
	}
	_, err = execKnownAction(t.Context(), nil, compose, planAction{Op: "composePs"})
	if err == nil || strings.Contains(err.Error(), "unknown op") {
		t.Fatalf("composePs: %v", err)
	}
	_, err = execKnownAction(t.Context(), nil, nil, planAction{Op: "start", fields: map[string]any{"id": "abc"}})
	if err == nil || strings.Contains(err.Error(), "unknown op") {
		t.Fatalf("start: %v", err)
	}
	_, err = execKnownAction(t.Context(), nil, nil, planAction{Op: "listImages"})
	if err == nil || strings.Contains(err.Error(), "unknown op") {
		t.Fatalf("listImages: %v", err)
	}

	var act planAction
	if err := act.UnmarshalJSON([]byte(`{"op":"port","service":"api","port":80,"repository":"app"}`)); err != nil {
		t.Fatal(err)
	}
	n, err := act.fieldInt("port")
	if err != nil || n != 80 || act.fieldString("service") != "api" || act.fieldString("repository") != "app" {
		t.Fatalf("fields = %#v port %d err %v", act.fields, n, err)
	}
}

func TestMarshalLooseReturnsMarshalError(t *testing.T) {
	t.Parallel()
	_, err := marshalLoose(make(chan int))
	if err == nil {
		t.Fatal("expected marshal error")
	}
}

func TestPlanValidateRejectsBadNumbers(t *testing.T) {
	t.Parallel()
	badRate := ImportPlan([]byte(`{"actions":[{"op":"run","options":{"blkioReadBps":[{"path":"/dev/sda","rate":"nope"}]}}]}`))
	if err := badRate.Validate(); err == nil {
		t.Fatal("expected rate decode error")
	}
	overflow := ImportPlan([]byte(`{"actions":[{"op":"run","options":{"blkioWeight":70000}}]}`))
	if err := overflow.Validate(); err == nil {
		t.Fatal("expected weight overflow error")
	}
}

func TestFieldIntRejectsMissingAndNonNumeric(t *testing.T) {
	t.Parallel()
	var act planAction
	if err := act.UnmarshalJSON([]byte(`{"op":"composePort","service":"api"}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := act.fieldInt("port"); err == nil {
		t.Fatal("expected missing port error")
	}
	if err := act.UnmarshalJSON([]byte(`{"op":"composePort","port":"nope"}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := act.fieldInt("port"); err == nil {
		t.Fatal("expected non-numeric port error")
	}
	if err := act.UnmarshalJSON([]byte(`{"op":"composePort","port":70000}`)); err != nil {
		t.Fatal(err)
	}
	n, err := act.fieldInt("port")
	if err != nil {
		t.Fatal(err)
	}
	if err := fitInt(n, "uint16"); err == nil {
		t.Fatal("expected uint16 overflow")
	}
}

func TestFlowHook(t *testing.T) {
	p := ImportPlan([]byte(`{"flow":{"name":"ci","steps":[]}}`))
	if err := p.Validate(); err == nil {
		t.Fatal("missing taskflow package")
	}
	if _, err := p.runActions(t.Context(), NewClient(nil, nil)); err == nil {
		t.Fatal("missing exec hook")
	}
	if ImportPlan([]byte(`{"flow":null}`)).hasFlow() {
		t.Fatal("null flow")
	}
	ValidateFlow = func(json.RawMessage) error { return nil }
	ExecFlow = func(context.Context, *Client, json.RawMessage) ([]byte, error) {
		return []byte(`{"ok":true}`), nil
	}
	t.Cleanup(func() {
		ValidateFlow = nil
		ExecFlow = nil
	})
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	out, err := p.runActions(t.Context(), NewClient(nil, nil))
	if err != nil || string(out) != `{"ok":true}` {
		t.Fatalf("%v %s", err, out)
	}
	both := ImportPlan([]byte(`{"flow":{"name":"ci","steps":[]},"actions":[{"op":"listImages"}]}`))
	if err := both.Validate(); err == nil || !strings.Contains(err.Error(), "flow and actions") {
		t.Fatalf("validate both: %v", err)
	}
	called := false
	ExecFlow = func(context.Context, *Client, json.RawMessage) ([]byte, error) {
		called = true
		return []byte(`{"ok":true}`), nil
	}
	if _, err := both.runActions(t.Context(), NewClient(nil, nil)); err == nil || !strings.Contains(err.Error(), "flow and actions") || called {
		t.Fatalf("exec both: %v called=%v", err, called)
	}
}

func TestDecodeRejectsBadTimeout(t *testing.T) {
	t.Parallel()
	_, err := decodeUpOptions(map[string]any{"timeout": "not-a-duration"})
	if err == nil {
		t.Fatal("expected error")
	}
}
