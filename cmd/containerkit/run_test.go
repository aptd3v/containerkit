package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func testdata(t *testing.T, name string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "tsx", "testdata", name)
}

func TestRunVersion(t *testing.T) {
	t.Parallel()
	var out, errw bytes.Buffer
	code := run(context.Background(), []string{"version"}, bytes.NewReader(nil), &out, &errw)
	if code != 0 {
		t.Fatalf("code = %d, err = %s", code, errw.String())
	}
	if !strings.HasPrefix(out.String(), "containerkit ") {
		t.Fatalf("out = %q", out.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	t.Parallel()
	var out, errw bytes.Buffer
	code := run(context.Background(), []string{"nope"}, bytes.NewReader(nil), &out, &errw)
	if code != 2 {
		t.Fatalf("code = %d", code)
	}
}

func TestRunValidatePlanJSON(t *testing.T) {
	t.Parallel()
	var out, errw bytes.Buffer
	code := run(context.Background(), []string{"validate", testdata(t, "plan.json")}, bytes.NewReader(nil), &out, &errw)
	if code != 0 {
		t.Fatalf("code = %d, err = %s", code, errw.String())
	}
}

func TestRunValidateYAML(t *testing.T) {
	t.Parallel()
	var out, errw bytes.Buffer
	code := run(context.Background(), []string{"validate", testdata(t, "example.yaml")}, bytes.NewReader(nil), &out, &errw)
	if code != 0 {
		t.Fatalf("code = %d, err = %s", code, errw.String())
	}
}

func TestRunValidateStdinJSON(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(testdata(t, "plan.json"))
	if err != nil {
		t.Fatal(err)
	}
	var out, errw bytes.Buffer
	code := run(context.Background(), []string{"validate", "-"}, bytes.NewReader(data), &out, &errw)
	if code != 0 {
		t.Fatalf("code = %d, err = %s", code, errw.String())
	}
}

func TestRunValidateBadJSON(t *testing.T) {
	t.Parallel()
	var out, errw bytes.Buffer
	code := run(context.Background(), []string{"validate", "-"}, strings.NewReader(`{`), &out, &errw)
	if code != 1 {
		t.Fatalf("code = %d, err = %s", code, errw.String())
	}
}

func TestRunSessionInvalidPlan(t *testing.T) {
	t.Parallel()
	var out, errw bytes.Buffer
	in := strings.NewReader("{\"id\":1,\"type\":\"exec\",\"plan\":{\"actions\":[{\"op\":\"\"}]}}\n")
	code := run(context.Background(), []string{"session"}, in, &out, &errw)
	if code != 0 {
		t.Fatalf("code = %d, err = %s", code, errw.String())
	}
	if !strings.Contains(out.String(), `"class":"invalidArgument"`) {
		t.Fatalf("out = %s, err = %s", out.String(), errw.String())
	}
}

func TestIsJSON(t *testing.T) {
	t.Parallel()
	if !isJSON([]byte("  {\n")) {
		t.Fatal("object")
	}
	if isJSON([]byte("name: app\n")) {
		t.Fatal("yaml")
	}
}
