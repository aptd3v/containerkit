package taskflow

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/aptd3v/containerkit/pkg/kit"
)

const appYAML = "name: app\nservices:\n  db:\n    image: postgres:16\n"

func TestImportRejects(t *testing.T) {
	cases := []struct {
		name string
		doc  string
	}{
		{"empty", ""},
		{"bad", "{"},
		{"null flow", `{"flow":null}`},
		{"empty name", `{"steps":[{"name":" ","op":"if","when":true}]}`},
		{"duplicate", `{"steps":[{"name":"a","op":"if","when":true},{"name":"a","op":"if","when":false}]}`},
		{"priority", `{"steps":[{"name":"a","op":"if","when":true,"priority":"nope"}]}`},
		{"task", `{"steps":[{"name":"a","op":"task"}]}`},
		{"condition", `{"steps":[{"name":"a","op":"condition"}]}`},
		{"if when", `{"steps":[{"name":"a","op":"if"}]}`},
		{"empty op", `{"steps":[{"name":"a"}]}`},
		{"depends", `{"steps":[{"name":"a","op":"if","when":true,"dependsOn":["missing"]}]}`},
		{"precede", `{"steps":[{"name":"a","op":"if","when":true,"precede":["missing"]}]}`},
		{"then name", `{"steps":[{"name":"a","op":"if","when":true,"then":["missing"]}]}`},
		{"else name", `{"steps":[{"name":"a","op":"if","when":true,"else":["missing"]}]}`},
		{"then json", `{"steps":[{"name":"a","op":"if","when":true,"then":{}}]}`},
		{"mixed", `{"steps":[{"name":"a","op":"if","when":true,"then":["b",{"name":"c","op":"if","when":false}]}]}`},
		{"build", `{"steps":[{"name":"a","op":"buildImage","options":{"noCache":"nope"}}]}`},
		{"container", `{"steps":[{"name":"a","op":"containerRun","options":{"autoRemove":"nope"}}]}`},
		{"compose yaml", `{"steps":[{"name":"a","op":"composeUp"}]}`},
		{"compose bad", `{"steps":[{"name":"a","op":"composeUp","yaml":"services: ["}]}`},
		{"subflow", `{"steps":[{"name":"a","op":"subflow","steps":[{"name":" ","op":"if"}]}]}`},
		{"else", `{"steps":[{"name":"a","op":"if","when":true,"else":{}}]}`},
		{"then value", `{"steps":[{"name":"a","op":"if","when":true,"then":[1]}]}`},
		{"flow shape", `{"flow":[]}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Import([]byte(tc.doc), nil); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestImportRuns(t *testing.T) {
	ctx := context.Background()
	doc := `{
	  "connect": {"fromEnv": true},
	  "flow": {
	    "name": "ci",
	    "steps": [
	      {"name": "build", "op": "buildImage", "options": {"parseDockerfile": "FROM alpine:latest\nCMD [\"true\"]\n", "tag": "app:ci"}},
	      {"name": "up-db", "op": "composeUp", "yaml": "name: app\nservices:\n  db:\n    image: postgres:16\n", "options": {"detach": true, "wait": true}, "dependsOn": ["build"]},
	      {"name": "test", "op": "containerRun", "options": {"image": "app:ci", "command": ["go", "test", "./..."]}, "dependsOn": ["up-db"], "retry": 1, "priority": "high"},
	      {"name": "release", "op": "if", "when": false, "dependsOn": ["test"], "else": [{"name": "note", "op": "tagImage", "options": {"source": "app:ci", "target": "app:ci"}}], "then": [{"name": "tag", "op": "tagImage", "options": {"source": "app:ci", "target": "app:release"}, "priority": "low", "always": true}]},
	      {"name": "inner", "op": "subflow", "steps": [{"name": "nested", "op": "if", "when": true, "priority": "normal"}], "precede": ["release"]}
	    ]
	  }
	}`
	eng := &scriptedEngine{logs: mux("ok\n"), buildBody: io.NopCloser(strings.NewReader("{\"stream\":\"ok\\n\"}\n"))}
	comp := &fakeComposer{}
	f, err := Import([]byte(doc), newClient(eng, comp))
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Run(ctx, discardRun()); err != nil {
		t.Fatal(err)
	}
	if comp.op != "up" {
		t.Fatal(comp.op)
	}

	plain := `{"name":"plain","steps":[{"name":"tag","op":"tagImage","options":{"source":"app","target":"app:v"},"service":"db","id":"1","path":"/","ref":"r","repository":"app","port":80}]}`
	f, err = Import([]byte(plain), newClient(&scriptedEngine{}, &fakeComposer{}))
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Run(ctx, discardRun()); err != nil {
		t.Fatal(err)
	}

	f, err = Import([]byte(`{"name":"no","steps":[{"name":"tag","op":"tagImage","options":{"source":"a","target":"b"}}]}`), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Run(ctx, discardRun()); !strings.Contains(err.Error(), ErrNoClient.Error()) && err == nil {
		t.Fatal(err)
	}

	f, err = Import([]byte(`{"name":"cup","steps":[{"name":"up","op":"composeUp","yaml":"`+yamlEscape(appYAML)+`"}]}`), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Run(ctx, discardRun()); err == nil {
		t.Fatal("nil client compose")
	}

	p := kit.ImportPlan([]byte(doc))
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	var wrapped struct {
		Flow json.RawMessage `json:"flow"`
	}
	if err := json.Unmarshal([]byte(doc), &wrapped); err != nil {
		t.Fatal(err)
	}
	out, err := kit.ExecFlow(ctx, newClient(&scriptedEngine{buildBody: io.NopCloser(strings.NewReader("{\"stream\":\"ok\\n\"}\n"))}, &fakeComposer{}), wrapped.Flow)
	if err != nil || string(out) != `{"ok":true}` {
		t.Fatalf("exec %v %s", err, out)
	}
	if _, err := kit.ExecFlow(ctx, newClient(&scriptedEngine{buildErr: errBuild()}, &fakeComposer{}), wrapped.Flow); err == nil {
		t.Fatal("exec failure")
	}
}

func TestImportUnknownOp(t *testing.T) {
	f, err := Import([]byte(`{"steps":[{"name":"a","op":"not-a-verb"}]}`), newClient(&scriptedEngine{}, &fakeComposer{}))
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Run(context.Background(), discardRun()); err == nil || !strings.Contains(err.Error(), "unknown op") {
		t.Fatal(err)
	}
}

func TestImportCycle(t *testing.T) {
	doc := `{"steps":[{"name":"a","op":"containerRun","dependsOn":["b"]},{"name":"b","op":"containerRun","dependsOn":["a"]}]}`
	_, err := Import([]byte(doc), nil)
	if err == nil || !cycleNames(err, "a", "b") {
		t.Fatal(err)
	}
}

func TestImportFlowFixture(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(file), "..", "..", "tsx", "testdata", "flow.json"))
	if err != nil {
		t.Fatal(err)
	}
	var plan struct {
		Actions []any           `json:"actions"`
		YAML    string          `json:"yaml"`
		Flow    json.RawMessage `json:"flow"`
	}
	if err := json.Unmarshal(data, &plan); err != nil {
		t.Fatal(err)
	}
	if len(plan.Actions) != 0 || plan.YAML != "" || len(bytes.TrimSpace(plan.Flow)) == 0 {
		t.Fatalf("plan envelope actions=%d yaml=%q flow=%d", len(plan.Actions), plan.YAML, len(plan.Flow))
	}
	f, err := Import(data, nil)
	if err != nil {
		t.Fatal(err)
	}
	if f.Name() != "ci" {
		t.Fatalf("name = %q", f.Name())
	}
	if err := kit.ImportPlan(data).Validate(); err != nil {
		t.Fatal(err)
	}
	body, err := flowBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	assertFlowFixture(t, body)
}

func assertFlowFixture(t *testing.T, body flowBody) {
	t.Helper()
	want := []string{"build", "lint", "up", "port", "down", "test", "gate", "ship", "hold", "note", "inner", "scan", "ping"}
	if body.Name != "ci" || len(body.Steps) != len(want) {
		t.Fatalf("steps = %d name %q", len(body.Steps), body.Name)
	}
	byName := map[string]stepDoc{}
	for i, st := range body.Steps {
		if st.Name != want[i] {
			t.Fatalf("steps[%d] = %q", i, st.Name)
		}
		byName[st.Name] = st
	}
	build := byName["build"]
	if build.Op != "buildImage" || build.Priority != "high" || len(build.DependsOn) != 0 || build.Options["noCache"] != false || build.Options["tag"] != "app:ci" {
		t.Fatalf("build = %+v", build)
	}
	env, ok := build.Options["env"].(map[string]any)
	labels, labelsOK := build.Options["label"].([]any)
	if !ok || env["value"] != "1" || !labelsOK || len(labels) != 1 {
		t.Fatalf("build env/label = %#v %#v", build.Options["env"], build.Options["label"])
	}
	lint := byName["lint"]
	if lint.Op != "containerRun" || lint.Priority != "normal" || lint.Always || lint.Retry == nil || *lint.Retry != 0 || len(lint.DependsOn) != 1 || lint.DependsOn[0] != "build" {
		t.Fatalf("lint = %+v", lint)
	}
	up := byName["up"]
	if up.Op != "composeUp" || !strings.Contains(up.YAML, "postgres:16") || up.Options["timeout"] != "30s" || up.Options["quiet"] != false {
		t.Fatalf("up = %+v", up.Options)
	}
	services, ok := up.Options["services"].([]any)
	if !ok || len(services) != 1 {
		t.Fatalf("up services = %#v", up.Options["services"])
	}
	port := byName["port"]
	if port.Op != "composePort" || port.Service != "db" || port.Port != 5432 || port.Options["protocol"] != "tcp" || port.Options["index"] != json.Number("0") {
		t.Fatalf("port = %+v %#v", port, port.Options)
	}
	down := byName["down"]
	if down.Op != "composeDown" || down.Options["services"] != "db" || down.Options["volumes"] != false || down.Options["timeout"] != json.Number("30000000000") {
		t.Fatalf("down = %#v", down.Options)
	}
	testStep := byName["test"]
	if testStep.Retry == nil || *testStep.Retry != 1 || testStep.Priority != "low" || len(testStep.DependsOn) != 2 {
		t.Fatalf("test = %+v", testStep)
	}
	gate := byName["gate"]
	if gate.Op != "if" || gate.When == nil || !*gate.When || string(bytes.TrimSpace(gate.Then)) != `["ship"]` || string(bytes.TrimSpace(gate.Else)) != `["hold"]` {
		t.Fatalf("gate then=%s else=%s when=%v", gate.Then, gate.Else, gate.When)
	}
	ship := byName["ship"]
	if ship.Service != "db" || ship.ID != "abc" || ship.Path != "compose.yaml" || ship.Ref != "app:ci" || ship.Repository != "app" || ship.Port != 80 {
		t.Fatalf("ship = %+v", ship)
	}
	if byName["hold"].Options == nil || len(byName["hold"].Options) != 0 || byName["hold"].Ref != "alpine:latest" {
		t.Fatalf("hold = %+v", byName["hold"])
	}
	note := byName["note"]
	if note.When == nil || *note.When || len(note.Precede) != 1 || note.Precede[0] != "gate" || string(bytes.TrimSpace(note.Then)) != "null" {
		t.Fatalf("note = %+v then=%s", note, note.Then)
	}
	_, nested, err := parseBranch(note.Else)
	if err != nil || len(nested) != 1 || nested[0].Name != "fallback" || !nested[0].Always || nested[0].Priority != "low" || nested[0].Retry == nil || *nested[0].Retry != 2 {
		t.Fatalf("note else = %+v %v", nested, err)
	}
	inner := byName["inner"]
	if inner.Op != "subflow" || len(inner.Precede) != 1 || inner.Precede[0] != "test" || len(inner.Steps) != 1 {
		t.Fatalf("inner = %+v", inner)
	}
	nestedIf := inner.Steps[0]
	if nestedIf.Name != "nested" || nestedIf.Op != "if" || nestedIf.When == nil || !*nestedIf.When || nestedIf.Priority != "normal" || string(bytes.TrimSpace(nestedIf.Then)) != "[]" {
		t.Fatalf("nested = %+v then=%s", nestedIf, nestedIf.Then)
	}
	_, skipped, err := parseBranch(nestedIf.Else)
	if err != nil || len(skipped) != 1 || skipped[0].Op != "listImages" || skipped[0].Options["all"] != "false" {
		t.Fatalf("skip = %+v %v", skipped, err)
	}
	filter, ok := skipped[0].Options["filter"].(map[string]any)
	if !ok || filter["values"] != "app:ci" {
		t.Fatalf("filter = %#v", skipped[0].Options["filter"])
	}
	scan := byName["scan"]
	filters, ok := scan.Options["filter"].([]any)
	if scan.Op != "listContainers" || !ok || len(filters) != 2 || scan.Options["limit"] != json.Number("10") || scan.Options["includeSize"] != false {
		t.Fatalf("scan = %#v", scan.Options)
	}
	ping := byName["ping"]
	if ping.Op != "ping" || ping.Options != nil || len(ping.DependsOn) != 3 {
		t.Fatalf("ping = %+v", ping)
	}
}

func yamlEscape(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\n", "\\n"), `"`, `\"`)
}

type buildErr string

func (e buildErr) Error() string { return string(e) }

func errBuild() error { return buildErr("build") }
