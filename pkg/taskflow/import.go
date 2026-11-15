package taskflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aptd3v/containerkit/pkg/kit"
)

func init() {
	kit.ValidateFlow = validateImported
	kit.ExecFlow = runImported
}

func validateImported(raw json.RawMessage) error {
	_, err := Import(raw, nil)
	return err
}

func runImported(ctx context.Context, cli *kit.Client, raw json.RawMessage) ([]byte, error) {
	f, err := Import(raw, cli)
	if err != nil {
		return nil, err
	}
	if err := f.Run(ctx, &Run{Writer: os.Stderr}); err != nil {
		return nil, err
	}
	return []byte(`{"ok":true}`), nil
}

type flowBody struct {
	Name  string    `json:"name"`
	Steps []stepDoc `json:"steps"`
}

type stepDoc struct {
	Name       string          `json:"name"`
	Op         string          `json:"op"`
	YAML       string          `json:"yaml"`
	Options    map[string]any  `json:"options"`
	DependsOn  []string        `json:"dependsOn"`
	Precede    []string        `json:"precede"`
	Then       json.RawMessage `json:"then"`
	Else       json.RawMessage `json:"else"`
	When       *bool           `json:"when"`
	Retry      *int            `json:"retry"`
	Always     bool            `json:"always"`
	Priority   string          `json:"priority"`
	Steps      []stepDoc       `json:"steps"`
	Service    string          `json:"service"`
	ID         string          `json:"id"`
	Path       string          `json:"path"`
	Ref        string          `json:"ref"`
	Repository string          `json:"repository"`
	Port       uint16          `json:"port"`
}

// Import builds a Flow from a flow document or from a plan document that
// carries a flow field. cli may be nil until Run.
func Import(data []byte, cli *kit.Client) (*Flow, error) {
	body, err := flowBytes(data)
	if err != nil {
		return nil, err
	}
	if err := validateSteps(body.Steps); err != nil {
		return nil, err
	}
	f := New(body.Name, cli)
	if _, err := addSteps(f.scope(), body.Steps, nil); err != nil {
		return nil, err
	}
	return f, nil
}

func flowBytes(data []byte) (flowBody, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return flowBody{}, fmt.Errorf("taskflow: empty document")
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var probe map[string]json.RawMessage
	if err := dec.Decode(&probe); err != nil {
		return flowBody{}, fmt.Errorf("taskflow: %w", err)
	}
	raw := data
	if flow, ok := probe["flow"]; ok {
		trimmed := bytes.TrimSpace(flow)
		if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
			return flowBody{}, fmt.Errorf("taskflow: empty flow")
		}
		raw = flow
	}
	dec = json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var body flowBody
	if err := dec.Decode(&body); err != nil {
		return flowBody{}, fmt.Errorf("taskflow: %w", err)
	}
	return body, nil
}

func validateSteps(steps []stepDoc) error {
	names := map[string]bool{}
	for i, st := range steps {
		if strings.TrimSpace(st.Name) == "" {
			return fmt.Errorf("taskflow: steps[%d]: name is empty", i)
		}
		if names[st.Name] {
			return fmt.Errorf("taskflow: duplicate step %q", st.Name)
		}
		names[st.Name] = true
		if err := prepare(st); err != nil {
			return fmt.Errorf("taskflow: step %s: %w", st.Name, err)
		}
		switch st.Op {
		case "subflow":
			if err := validateSteps(st.Steps); err != nil {
				return err
			}
		case "if":
			if err := validateBranch(st.Then); err != nil {
				return fmt.Errorf("taskflow: step %s: then: %w", st.Name, err)
			}
			if err := validateBranch(st.Else); err != nil {
				return fmt.Errorf("taskflow: step %s: else: %w", st.Name, err)
			}
		}
	}
	for _, st := range steps {
		if err := knownEdges(st.Name, st.DependsOn, names, "depends on"); err != nil {
			return err
		}
		if err := knownEdges(st.Name, st.Precede, names, "precedes"); err != nil {
			return err
		}
		if st.Op != "if" {
			continue
		}
		thenNames, _, _ := parseBranch(st.Then)
		elseNames, _, _ := parseBranch(st.Else)
		if err := knownEdges(st.Name, thenNames, names, "then"); err != nil {
			return err
		}
		if err := knownEdges(st.Name, elseNames, names, "else"); err != nil {
			return err
		}
	}
	return nil
}

func knownEdges(step string, edges []string, names map[string]bool, kind string) error {
	for _, name := range edges {
		if !names[name] {
			return fmt.Errorf("taskflow: step %s %s unknown step %q", step, kind, name)
		}
	}
	return nil
}

func validateBranch(raw json.RawMessage) error {
	names, nested, err := parseBranch(raw)
	if err != nil {
		return err
	}
	if len(names) > 0 && len(nested) > 0 {
		return fmt.Errorf("branch mixes step names and step objects")
	}
	return validateSteps(nested)
}

func prepare(st stepDoc) error {
	switch st.Priority {
	case "", "high", "normal", "low":
	default:
		return fmt.Errorf("priority %q", st.Priority)
	}
	switch st.Op {
	case "":
		return fmt.Errorf("op is empty")
	case "task", "condition":
		return fmt.Errorf("%s callbacks are Go-only", st.Op)
	case "if":
		if st.When == nil {
			return fmt.Errorf("if requires when")
		}
		return nil
	case "subflow":
		return nil
	case "buildImage", "containerRun":
		return nil
	default:
		if strings.HasPrefix(st.Op, "compose") && strings.TrimSpace(st.YAML) == "" {
			return fmt.Errorf("yaml required for %s", st.Op)
		}
		if strings.HasPrefix(st.Op, "compose") {
			if err := kit.ImportYAML([]byte(st.YAML)).Validate(); err != nil {
				return err
			}
		}
		return kit.DecodeAction(st.Op, st.Options)
	}
}

func addSteps(sc scope, steps []stepDoc, known map[string]*Task) (map[string]*Task, error) {
	made := map[string]*Task{}
	var ordered []stepDoc
	for _, st := range steps {
		t, err := makeStep(sc, st)
		if err != nil {
			return nil, err
		}
		applyProps(t, st)
		made[st.Name] = t
		ordered = append(ordered, st)
	}
	lookup := func(name string) *Task {
		if t := made[name]; t != nil {
			return t
		}
		return known[name]
	}
	for _, st := range ordered {
		t := made[st.Name]
		var deps []*Task
		for _, name := range st.DependsOn {
			deps = append(deps, lookup(name))
		}
		if len(deps) > 0 {
			t.DependsOn(deps...)
		}
		var next []*Task
		for _, name := range st.Precede {
			next = append(next, lookup(name))
		}
		if len(next) > 0 {
			t.Precede(next...)
		}
		if st.Op != "if" {
			continue
		}
		if err := wireBranch(sc, t, st.Then, true, made, known); err != nil {
			return nil, err
		}
		if err := wireBranch(sc, t, st.Else, false, made, known); err != nil {
			return nil, err
		}
	}
	return made, nil
}

func wireBranch(sc scope, t *Task, raw json.RawMessage, then bool, made, known map[string]*Task) error {
	names, nested, err := parseBranch(raw)
	if err != nil {
		return err
	}
	kind := "else"
	if then {
		kind = "then"
	}
	var tasks []*Task
	for _, name := range names {
		task := made[name]
		if task == nil {
			task = known[name]
		}
		if task == nil {
			return fmt.Errorf("taskflow: step %s %s unknown step %q", t.name, kind, name)
		}
		tasks = append(tasks, task)
	}
	if len(nested) > 0 {
		extra, err := addSteps(sc, nested, mergeTasks(known, made))
		if err != nil {
			return err
		}
		for _, st := range nested {
			tasks = append(tasks, extra[st.Name])
		}
	}
	if len(tasks) == 0 {
		return nil
	}
	if then {
		t.Then(tasks...)
	} else {
		t.Else(tasks...)
	}
	return nil
}

func mergeTasks(a, b map[string]*Task) map[string]*Task {
	out := map[string]*Task{}
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

func makeStep(sc scope, st stepDoc) (*Task, error) {
	switch st.Op {
	case "buildImage":
		opts, err := kit.EngineBuildOptions(st.Options)
		if err != nil {
			return nil, err
		}
		return sc.build(st.Name, opts), nil
	case "containerRun":
		ctr, err := kit.ContainerFromOptions(st.Options)
		if err != nil {
			return nil, err
		}
		return sc.container(st.Name, ctr, nil), nil
	case "if":
		when := false
		if st.When != nil {
			when = *st.When
		}
		return sc.ifStep(st.Name, func(context.Context) bool { return when }), nil
	case "subflow":
		nested := append([]stepDoc(nil), st.Steps...)
		return sc.subflow(st.Name, importSubflow(st.Name, nested)), nil
	default:
		fields := actionFields(st)
		op := st.Op
		options := st.Options
		var app *kit.Compose
		if strings.HasPrefix(op, "compose") {
			app = kit.ImportYAML([]byte(st.YAML))
		}
		fn := func(ctx context.Context, app *kit.Compose, _ kit.Composer) error {
			if sc.flow.cli == nil {
				return ErrNoClient
			}
			_, err := kit.ExecAction(ctx, sc.flow.cli, app, op, options, fields)
			return err
		}
		if app != nil {
			return sc.compose(st.Name, app, fn), nil
		}
		return sc.task(st.Name, func(ctx context.Context) error {
			return fn(ctx, nil, nil)
		}), nil
	}
}

func actionFields(st stepDoc) map[string]any {
	fields := map[string]any{}
	if st.Service != "" {
		fields["service"] = st.Service
	}
	if st.ID != "" {
		fields["id"] = st.ID
	}
	if st.Path != "" {
		fields["path"] = st.Path
	}
	if st.Ref != "" {
		fields["ref"] = st.Ref
	}
	if st.Repository != "" {
		fields["repository"] = st.Repository
	}
	if st.Port != 0 {
		fields["port"] = int(st.Port)
	}
	return fields
}

func applyProps(t *Task, st stepDoc) {
	if st.Retry != nil {
		t.Retry(*st.Retry)
	}
	if st.Always {
		t.Always()
	}
	switch st.Priority {
	case "high":
		t.Priority(PriorityHigh)
	case "normal":
		t.Priority(PriorityNormal)
	case "low":
		t.Priority(PriorityLow)
	}
}

func importSubflow(name string, nested []stepDoc) func(*Subflow) error {
	return func(sf *Subflow) error {
		inner, ok := sf.scope()
		if !ok {
			return fmt.Errorf("taskflow: subflow %s: scope is not ready", name)
		}
		_, err := addSteps(inner, nested, nil)
		return err
	}
}

func parseBranch(raw json.RawMessage) (names []string, nested []stepDoc, err error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil, nil, nil
	}
	dec := json.NewDecoder(bytes.NewReader(trimmed))
	dec.UseNumber()
	var arr []json.RawMessage
	if err := dec.Decode(&arr); err != nil {
		return nil, nil, err
	}
	return parseBranchElements(arr)
}

func parseBranchElements(arr []json.RawMessage) (names []string, nested []stepDoc, err error) {
	for _, el := range arr {
		el, err = nonemptyBranchElement(el)
		if err != nil {
			return nil, nil, err
		}
		if el[0] == '"' {
			var name string
			_ = json.Unmarshal(el, &name)
			names = append(names, name)
			continue
		}
		dec := json.NewDecoder(bytes.NewReader(el))
		dec.UseNumber()
		var st stepDoc
		if err := dec.Decode(&st); err != nil {
			return nil, nil, err
		}
		nested = append(nested, st)
	}
	return names, nested, nil
}

func nonemptyBranchElement(el []byte) ([]byte, error) {
	el = bytes.TrimSpace(el)
	if len(el) == 0 {
		return nil, fmt.Errorf("taskflow: empty branch element")
	}
	return el, nil
}
