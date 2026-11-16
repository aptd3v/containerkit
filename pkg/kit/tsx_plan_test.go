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
	if p.hasFlow() {
		t.Fatal("flow is set beside actions")
	}
	assertPlanConnect(t, p.raw.Connect)
	assertPlanActions(t, p.raw.Actions)
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
	cfg, err := svcs[0].ComposeConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Command) != 6 || cfg.Command[0] != "httpd" || cfg.User != "65534:65534" || cfg.WorkingDir != "/www" || cfg.Restart != "unless-stopped" {
		t.Fatalf("command/user/workdir/restart = %+v %q %q %q", cfg.Command, cfg.User, cfg.WorkingDir, cfg.Restart)
	}
	if cfg.HealthCheck == nil || cfg.HealthCheck.Retries == nil || *cfg.HealthCheck.Retries != 3 || cfg.HealthCheck.StartPeriod == nil {
		t.Fatalf("health = %+v", cfg.HealthCheck)
	}
	if len(cfg.Ports) != 1 || cfg.Ports[0].Published != "8080" || cfg.Ports[0].Target != 80 || cfg.Ports[0].Protocol != "tcp" {
		t.Fatalf("ports = %+v", cfg.Ports)
	}
	if cfg.Environment["APP"] == nil || *cfg.Environment["APP"] != "1" || cfg.Environment["EMPTY"] == nil || *cfg.Environment["EMPTY"] != "" {
		t.Fatalf("env = %+v", cfg.Environment)
	}
	if cfg.Labels["kit.edge"] != "yes" {
		t.Fatalf("labels = %+v", cfg.Labels)
	}
	if cfg.Logging == nil || cfg.Logging.Driver != "json-file" || cfg.Logging.Options["max-size"] != "1m" {
		t.Fatalf("logging = %+v", cfg.Logging)
	}
	if cfg.Deploy == nil || cfg.Deploy.Replicas == nil || *cfg.Deploy.Replicas != 1 {
		t.Fatalf("deploy = %+v", cfg.Deploy)
	}
	if len(cfg.Volumes) != 1 || cfg.Volumes[0].Source != "data" || cfg.Volumes[0].Target != "/data" || !cfg.Volumes[0].ReadOnly {
		t.Fatalf("volumes = %+v", cfg.Volumes)
	}
	if cfg.Networks["front"] == nil {
		t.Fatalf("networks = %+v", cfg.Networks)
	}
}

func assertPlanConnect(t *testing.T, connect map[string]any) {
	t.Helper()
	if connect["fromEnv"] != true || connect["hostFromEnv"] != false || connect["dryRun"] != false || connect["tlsFromEnv"] != false || connect["apiVersionFromEnv"] != false {
		t.Fatalf("bools = %#v", connect)
	}
	if connect["negotiateAPIVersion"] != true {
		t.Fatalf("negotiateAPIVersion = %#v", connect["negotiateAPIVersion"])
	}
	if connect["host"] != "unix:///var/run/docker.sock" || connect["scheme"] != "unix" || connect["apiVersion"] != "1.44" || connect["timeout"] != "10s" || connect["userAgent"] != "containerkit-tsx-test" {
		t.Fatalf("strings = %#v", connect)
	}
	if connect["maxConcurrency"] != json.Number("4") {
		t.Fatalf("maxConcurrency = %#v (%T)", connect["maxConcurrency"], connect["maxConcurrency"])
	}
	tls, ok := connect["tls"].(map[string]any)
	if !ok || tls["ca"] != "/ca.pem" || tls["cert"] != "/cert.pem" || tls["key"] != "/key.pem" {
		t.Fatalf("tls = %#v", connect["tls"])
	}
	headers, ok := connect["httpHeader"].([]any)
	if !ok || len(headers) != 2 {
		t.Fatalf("httpHeader = %#v", connect["httpHeader"])
	}
	last, ok := headers[1].(map[string]any)
	if !ok || last["name"] != "X-Empty" || last["value"] != "" {
		t.Fatalf("empty header = %#v", headers[1])
	}
}

func assertPlanActions(t *testing.T, actions []planAction) {
	t.Helper()
	want := []string{
		"pullImage",
		"listImages",
		"listContainers",
		"inspectImage",
		"inspect",
		"load",
		"composePs",
		"composePort",
		"composeKill",
		"composeUp",
		"composeDown",
		"ping",
		"composeVolumeList",
	}
	if len(actions) != len(want) {
		t.Fatalf("actions = %d", len(actions))
	}
	byOp := map[string]planAction{}
	for i, act := range actions {
		if act.Op != want[i] {
			t.Fatalf("actions[%d] = %q", i, act.Op)
		}
		byOp[act.Op] = act
	}
	pull := byOp["pullImage"]
	if pull.fieldString("ref") != "alpine:latest" || pull.Options["all"] != true || pull.Options["platform"] != "linux/arm64" {
		t.Fatalf("pullImage = %#v", pull)
	}
	images := byOp["listImages"]
	if images.Options["all"] != "false" {
		t.Fatalf("listImages all = %#v (%T)", images.Options["all"], images.Options["all"])
	}
	oneFilter, ok := images.Options["filter"].(map[string]any)
	if !ok || oneFilter["term"] != "reference" || oneFilter["values"] != "alpine*" {
		t.Fatalf("single filter = %#v", images.Options["filter"])
	}
	containers := byOp["listContainers"]
	filters, ok := containers.Options["filter"].([]any)
	if !ok || len(filters) != 2 || containers.Options["limit"] != json.Number("10") || containers.Options["includeSize"] != false {
		t.Fatalf("listContainers = %#v", containers.Options)
	}
	first, ok := filters[0].(map[string]any)
	values, valuesOK := first["values"].([]any)
	if !ok || !valuesOK || len(values) != 2 {
		t.Fatalf("filter values = %#v", filters[0])
	}
	inspected := byOp["inspectImage"]
	platform, ok := inspected.Options["imagePlatform"].(map[string]any)
	features, featuresOK := platform["os.features"].([]any)
	if !ok || !featuresOK || platform["os.version"] != "12" || platform["variant"] != "v8" || len(features) != 1 || inspected.Options["manifests"] != false {
		t.Fatalf("imagePlatform = %#v", inspected.Options["imagePlatform"])
	}
	if byOp["inspect"].fieldString("id") != "abc" || byOp["inspect"].Options["includeSize"] != false {
		t.Fatalf("inspect = %#v", byOp["inspect"])
	}
	loaded := byOp["load"]
	envFile, ok := loaded.Options["envFile"].(map[string]any)
	services, servicesOK := loaded.Options["services"].([]any)
	if loaded.fieldString("path") != "compose.yaml" || !ok || envFile["path"] != ".env" || envFile["required"] != false || !servicesOK || len(services) != 1 {
		t.Fatalf("load = %#v", loaded)
	}
	if byOp["composePs"].Options["services"] != "api" || byOp["composePs"].Options["all"] != true {
		t.Fatalf("composePs = %#v", byOp["composePs"].Options)
	}
	port := byOp["composePort"]
	if port.fieldString("service") != "api" || port.fields["port"] != json.Number("80") || port.Options["protocol"] != "tcp" || port.Options["index"] != "0" {
		t.Fatalf("composePort = %#v fields %#v", port.Options, port.fields)
	}
	killServices, ok := byOp["composeKill"].Options["services"].([]any)
	if !ok || len(killServices) != 1 || byOp["composeKill"].Options["stopSignal"] != "SIGTERM" || byOp["composeKill"].Options["all"] != false {
		t.Fatalf("composeKill = %#v", byOp["composeKill"].Options)
	}
	up := byOp["composeUp"].Options
	if up["detach"] != true || up["timeout"] != "30s" || up["quiet"] != false || up["progress"] != "plain" || up["memory"] != json.Number("536870912") {
		t.Fatalf("composeUp = %#v", up)
	}
	down := byOp["composeDown"].Options
	if down["timeout"] != json.Number("30000000000") || down["volumes"] != false || down["services"] != "api" || down["removeOrphans"] != true {
		t.Fatalf("composeDown = %#v", down)
	}
	if byOp["ping"].Options != nil {
		t.Fatalf("ping options = %#v", byOp["ping"].Options)
	}
	if byOp["composeVolumeList"].Options == nil || len(byOp["composeVolumeList"].Options) != 0 {
		t.Fatalf("composeVolumeList = %#v", byOp["composeVolumeList"].Options)
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
