package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestComposeStepSpecPairsOps(t *testing.T) {
	var buf bytes.Buffer
	writeComposeStepSpec(&buf, []verb{
		{op: "pullImage", bag: "PullImageOptions"},
		{op: "composeUp", bag: "UpOptions"},
		{
			op:          "composePort",
			bag:         "ComposePortOptions",
			positionals: []leadParam{{name: "service", tsType: "string"}, {name: "port", tsType: "number"}},
		},
		{
			op:          "composePublishOCI",
			bag:         "PublishOptions",
			positionals: []leadParam{{name: "repository", tsType: "string"}},
		},
		{op: "composeImages"},
	})
	got := buf.String()
	for _, want := range []string{
		`| { op: "composeUp"; options?: UpOptions }`,
		`| { op: "composePort"; service: string; port: number; options?: ComposePortOptions }`,
		`| { op: "composePublishOCI"; repository: string; options?: PublishOptions }`,
		`| { op: "composeImages" }`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %s\n%s", want, got)
		}
	}
	if strings.Contains(got, "pullImage") {
		t.Fatal(got)
	}
}
