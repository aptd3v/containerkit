package taskflow

import (
	"context"
	"encoding/json"
	"io"
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

func yamlEscape(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\n", "\\n"), `"`, `\"`)
}

type buildErr string

func (e buildErr) Error() string { return string(e) }

func errBuild() error { return buildErr("build") }
