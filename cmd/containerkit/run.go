package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/aptd3v/containerkit/pkg/kit"
	_ "github.com/aptd3v/containerkit/pkg/taskflow"
)

const usage = `containerkit — run a containerkit plan or compose YAML against a Docker daemon.

Usage:
  containerkit version
  containerkit validate [file]
  containerkit exec [file]
  containerkit session

file is a path, omitted, or "-" for stdin.
JSON is a TSX plan (encode(composeUp(...))). YAML is a compose file (connect from env, then composeUp).
validate does not dial the daemon. exec does.
`

func run(ctx context.Context, args []string, in io.Reader, out, errw io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(errw, usage)
		return 2
	}
	switch args[0] {
	case "version", "-version", "--version":
		fmt.Fprintf(out, "containerkit %s (%s)\n", version, commit)
		return 0
	case "help", "-h", "--help":
		fmt.Fprint(out, usage)
		return 0
	case "validate":
		data, err := readInput(args[1:], in)
		if err != nil {
			fmt.Fprintln(errw, err)
			return 1
		}
		if err := validate(data); err != nil {
			fmt.Fprintln(errw, err)
			return 1
		}
		return 0
	case "exec":
		data, err := readInput(args[1:], in)
		if err != nil {
			fmt.Fprintln(errw, err)
			return 1
		}
		result, err := execPlan(ctx, data)
		if err != nil {
			fmt.Fprintln(errw, err)
			return 1
		}
		if _, err := out.Write(result); err != nil {
			fmt.Fprintln(errw, err)
			return 1
		}
		if len(result) > 0 && result[len(result)-1] != '\n' {
			fmt.Fprintln(out)
		}
		return 0
	case "session":
		return kit.ServeSession(ctx, in, out, errw)
	default:
		fmt.Fprintf(errw, "unknown command %q\n%s", args[0], usage)
		return 2
	}
}

func readInput(args []string, in io.Reader) ([]byte, error) {
	path := ""
	if len(args) > 0 {
		path = args[0]
	}
	if path == "" || path == "-" {
		return io.ReadAll(in)
	}
	return os.ReadFile(path)
}

func validate(data []byte) error {
	if isJSON(data) {
		return kit.ImportPlan(data).Validate()
	}
	return kit.ImportYAML(data).Validate()
}

func execPlan(ctx context.Context, data []byte) ([]byte, error) {
	if isJSON(data) {
		return kit.ExecJSON(ctx, data)
	}
	wrapped, err := json.Marshal(map[string]any{
		"connect": map[string]any{"fromEnv": true},
		"yaml":    string(data),
		"actions": []map[string]any{{"op": "composeUp"}},
	})
	if err != nil {
		return nil, err
	}
	return kit.ExecJSON(ctx, wrapped)
}

func isJSON(data []byte) bool {
	s := bytes.TrimSpace(data)
	return len(s) > 0 && (s[0] == '{' || s[0] == '[')
}
