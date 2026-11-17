package kit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aptd3v/containerkit/pkg/progress"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// Plan is a TSX-encoded connect + compose YAML + action document.
// Construct with [ImportPlan]. Decode errors accumulate; call [Plan.Validate]
// before [Plan.Run] / [ExecJSON].
type Plan struct {
	raw  jsonPlan
	errs []error
}

type jsonPlan struct {
	Connect  map[string]any  `json:"connect"`
	YAML     string          `json:"yaml"`
	Actions  []planAction    `json:"actions"`
	Flow     json.RawMessage `json:"flow"`
	Progress string          `json:"progress"`
}

// ExecFlow runs a taskflow document on an already-dialed client.
// Package taskflow registers it. A nil hook means that package is not linked.
var ExecFlow func(ctx context.Context, cli *Client, flow json.RawMessage) ([]byte, error)

// ValidateFlow checks a taskflow document without dialing. Package taskflow
// registers it. A nil hook means that package is not linked.
var ValidateFlow func(flow json.RawMessage) error

type planAction struct {
	Op      string         `json:"op"`
	Ref     string         `json:"ref"`
	Service string         `json:"service"`
	ID      string         `json:"id"`
	Path    string         `json:"path"`
	Options map[string]any `json:"options"`
	fields  map[string]any
}

func (a *planAction) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var m map[string]any
	if err := dec.Decode(&m); err != nil {
		return err
	}
	a.fields = m
	a.Op, _ = asString(m["op"])
	a.Ref, _ = asString(m["ref"])
	a.Service, _ = asString(m["service"])
	a.ID, _ = asString(m["id"])
	a.Path, _ = asString(m["path"])
	if raw, ok := m["options"]; ok && raw != nil {
		obj, err := asObject(raw)
		if err != nil {
			return err
		}
		a.Options = obj
	}
	return nil
}

func (a planAction) fieldString(key string) string {
	if a.fields != nil {
		s, _ := asString(a.fields[key])
		return s
	}
	switch key {
	case "ref":
		return a.Ref
	case "service":
		return a.Service
	case "id":
		return a.ID
	case "path":
		return a.Path
	default:
		return ""
	}
}

func (a planAction) fieldInt(key string) (int, error) {
	if a.fields == nil {
		return 0, fmt.Errorf("%w: %s", errRequired, key)
	}
	v, ok := a.fields[key]
	if !ok || v == nil {
		return 0, fmt.Errorf("%w: %s", errRequired, key)
	}
	return asInt(v)
}

// ImportPlan loads a TSX plan JSON document. Parse errors accumulate on the
// returned [Plan]; call [Plan.Validate].
func ImportPlan(data []byte) *Plan {
	p := &Plan{}
	if len(data) == 0 {
		p.errs = append(p.errs, cfgErr("tsx", "plan", "ImportPlan", fmt.Errorf("%w: json", errEmpty)))
		return p
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&p.raw); err != nil {
		p.errs = append(p.errs, cfgErr("tsx", "plan", "ImportPlan", err))
	}
	return p
}

func (p *Plan) hasFlow() bool {
	if p == nil {
		return false
	}
	raw := bytes.TrimSpace(p.raw.Flow)
	return len(raw) > 0 && !bytes.Equal(raw, []byte("null"))
}

// Validate decodes connect/action option bags and, when YAML is present,
// validates the compose file. It does not dial the daemon.
func (p *Plan) Validate() error {
	if p == nil {
		return cfgErr("tsx", "plan", "Validate", fmt.Errorf("%w: plan", errRequired))
	}
	var errs []error
	errs = append(errs, p.errs...)
	if _, err := decodeConnectOptions(p.raw.Connect); err != nil {
		errs = append(errs, err)
	}
	if p.hasFlow() {
		if len(p.raw.Actions) > 0 {
			errs = append(errs, cfgErr("tsx", "plan", "Validate", fmt.Errorf("%w: flow and actions", errInvalid)))
		}
		if ValidateFlow == nil {
			errs = append(errs, cfgErr("tsx", "flow", "Validate", fmt.Errorf("%w: package taskflow", errRequired)))
		} else if err := ValidateFlow(p.raw.Flow); err != nil {
			errs = append(errs, err)
		}
	}
	for i, act := range p.raw.Actions {
		if act.Op == "" {
			errs = append(errs, cfgErr("tsx", fmt.Sprintf("actions[%d]", i), "op", fmt.Errorf("%w: op", errEmpty)))
			continue
		}
		if err := decodeActionOptions(act.Op, act.Options); err != nil {
			errs = append(errs, fmt.Errorf("actions[%d].%s: %w", i, act.Op, err))
		}
	}
	if p.raw.YAML != "" {
		c := ImportYAML([]byte(p.raw.YAML))
		if err := c.Validate(); err != nil {
			errs = append(errs, err)
		}
	}
	if _, err := progressMode(p.raw.Progress); err != nil {
		errs = append(errs, err)
	}
	return join(errs...)
}

// Compose loads the plan YAML through [ImportYAML].
func (p *Plan) Compose() *Compose {
	if p == nil || p.raw.YAML == "" {
		return importErr(cfgErr("tsx", "plan", "Compose", fmt.Errorf("%w: yaml", errEmpty)))
	}
	return ImportYAML([]byte(p.raw.YAML))
}

// Run dials [Connect] from the plan and executes its actions.
// Progress is written only when the plan sets progress. An empty field stays silent.
func (p *Plan) Run(ctx context.Context) error {
	_, err := p.exec(ctx)
	return err
}

// ExecJSON loads a plan and runs it, returning the last action's JSON result.
// Progress is written to stderr only when the plan sets progress.
func ExecJSON(ctx context.Context, data []byte) ([]byte, error) {
	return ImportPlan(data).exec(ctx)
}

// ExecCLI runs a plan the way containerkit exec does.
// Progress is written to errw. An empty progress field uses auto.
// "off" writes nothing. The returned bytes are the action result.
func ExecCLI(ctx context.Context, data []byte, errw io.Writer) ([]byte, error) {
	return ImportPlan(data).execProgress(ctx, errw, true)
}

func (p *Plan) exec(ctx context.Context) ([]byte, error) {
	return p.execProgress(ctx, nil, false)
}

func progressMode(field string) (string, error) {
	mode := strings.ToLower(strings.TrimSpace(field))
	switch mode {
	case "", "auto", "plain", "tty", "quiet", "json", "off":
		return mode, nil
	default:
		return "", cfgErr("tsx", "progress", "progress", fmt.Errorf("%w: progress %q", errInvalid, field))
	}
}

// attachPlanProgress puts a logger on ctx. An empty field stays silent unless
// defaultAuto is set, which is the containerkit exec command. "off" writes nothing.
func attachPlanProgress(ctx context.Context, field string, errw io.Writer, defaultAuto bool) (context.Context, error) {
	mode, err := progressMode(field)
	if err != nil {
		return ctx, err
	}
	if mode == "" && defaultAuto {
		mode = "auto"
	}
	if mode == "" || mode == "off" {
		return ctx, nil
	}
	w := errw
	if w == nil {
		w = os.Stderr
	}
	return NewLogger(OutputStream(w), Progress(progress.Mode(mode))).Context(ctx), nil
}

func (p *Plan) execProgress(ctx context.Context, errw io.Writer, defaultAuto bool) ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	ctx, err := attachPlanProgress(ctx, p.raw.Progress, errw, defaultAuto)
	if err != nil {
		return nil, err
	}
	opts, err := decodeConnectOptions(p.raw.Connect)
	if err != nil {
		return nil, err
	}
	cli, err := Connect(opts...)
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	return p.runActions(ctx, cli)
}

// runActions executes a validated plan on an existing client. The caller owns
// the client and decides when it is closed.
func (p *Plan) runActions(ctx context.Context, cli *Client) ([]byte, error) {
	if cli == nil {
		return nil, cfgErr("tsx", "plan", "exec", fmt.Errorf("%w: client", errRequired))
	}
	if p.hasFlow() {
		if len(p.raw.Actions) > 0 {
			return nil, cfgErr("tsx", "plan", "exec", fmt.Errorf("%w: flow and actions", errInvalid))
		}
		if ExecFlow == nil {
			return nil, cfgErr("tsx", "flow", "exec", fmt.Errorf("%w: package taskflow", errRequired))
		}
		return ExecFlow(ctx, cli, p.raw.Flow)
	}
	var compose *Compose
	if p.raw.YAML != "" {
		compose = ImportYAML([]byte(p.raw.YAML))
	}
	var last []byte
	for _, act := range p.raw.Actions {
		var err error
		last, err = execKnownAction(ctx, cli, compose, act)
		if err != nil {
			return nil, err
		}
	}
	if last == nil {
		last = []byte(`{"ok":true}`)
	}
	return last, nil
}

func marshalLoose(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("tsx plan: marshal result: %w", err)
	}
	return b, nil
}

func asObject(v any) (map[string]any, error) {
	switch t := v.(type) {
	case nil:
		return map[string]any{}, nil
	case map[string]any:
		return t, nil
	default:
		return nil, fmt.Errorf("expected object, got %T", v)
	}
}

func asObjectList(v any) ([]map[string]any, error) {
	switch t := v.(type) {
	case nil:
		return nil, nil
	case map[string]any:
		return []map[string]any{t}, nil
	case []any:
		out := make([]map[string]any, 0, len(t))
		for i, item := range t {
			m, err := asObject(item)
			if err != nil {
				return nil, fmt.Errorf("[%d]: %w", i, err)
			}
			out = append(out, m)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("expected object or array, got %T", v)
	}
}

func asBool(v any) (bool, error) {
	switch t := v.(type) {
	case bool:
		return t, nil
	case string:
		return strconv.ParseBool(t)
	default:
		return false, fmt.Errorf("expected bool, got %T", v)
	}
}

func asPlatform(v any) (ocispec.Platform, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return ocispec.Platform{}, err
	}
	var p ocispec.Platform
	if err := json.Unmarshal(raw, &p); err != nil {
		return ocispec.Platform{}, err
	}
	return p, nil
}

func asString(v any) (string, error) {
	switch t := v.(type) {
	case nil:
		return "", nil
	case string:
		return t, nil
	case json.Number:
		return t.String(), nil
	default:
		return fmt.Sprint(t), nil
	}
}

func asInt(v any) (int, error) {
	switch t := v.(type) {
	case json.Number:
		n, err := t.Int64()
		if err != nil {
			return 0, err
		}
		if int64(int(n)) != n {
			return 0, fmt.Errorf("%w: %d overflows int", errInvalid, n)
		}
		return int(n), nil
	case float64:
		if math.Trunc(t) != t {
			return 0, fmt.Errorf("%w: %v is not an integer", errInvalid, t)
		}
		if t > float64(math.MaxInt) || t < float64(math.MinInt) {
			return 0, fmt.Errorf("%w: %v overflows int", errInvalid, t)
		}
		return int(t), nil
	case int:
		return t, nil
	case string:
		n, err := strconv.Atoi(t)
		return n, err
	default:
		return 0, fmt.Errorf("expected number, got %T", v)
	}
}

func asDuration(v any) (time.Duration, error) {
	switch t := v.(type) {
	case string:
		return time.ParseDuration(t)
	case json.Number:
		n, err := t.Int64()
		return time.Duration(n), err
	case float64:
		return time.Duration(t), nil
	default:
		return 0, fmt.Errorf("expected duration, got %T", v)
	}
}

func asStringSlice(v any) ([]string, error) {
	switch t := v.(type) {
	case nil:
		return nil, nil
	case string:
		return []string{t}, nil
	case []any:
		out := make([]string, 0, len(t))
		for _, item := range t {
			s, err := asString(item)
			if err != nil {
				return nil, err
			}
			out = append(out, s)
		}
		return out, nil
	default:
		s, err := asString(v)
		if err != nil {
			return nil, err
		}
		return []string{s}, nil
	}
}

func protoSlice(ss []string) []Proto {
	out := make([]Proto, len(ss))
	for i, s := range ss {
		out[i] = Proto(s)
	}
	return out
}

func fitInt(n int, goType string) error {
	switch goType {
	case "", "int", "int64":
		return nil
	case "int8":
		if n < math.MinInt8 || n > math.MaxInt8 {
			return overflowInt(n, goType)
		}
	case "int16":
		if n < math.MinInt16 || n > math.MaxInt16 {
			return overflowInt(n, goType)
		}
	case "int32":
		if n < math.MinInt32 || n > math.MaxInt32 {
			return overflowInt(n, goType)
		}
	case "uint8":
		if n < 0 || n > math.MaxUint8 {
			return overflowInt(n, goType)
		}
	case "uint16":
		if n < 0 || n > math.MaxUint16 {
			return overflowInt(n, goType)
		}
	case "uint32":
		if n < 0 || int64(n) > math.MaxUint32 {
			return overflowInt(n, goType)
		}
	case "uint", "uint64":
		if n < 0 {
			return overflowInt(n, goType)
		}
	}
	return nil
}

func overflowInt(n int, goType string) error {
	return fmt.Errorf("%w: %d overflows %s", errInvalid, n, goType)
}

func firstKey(m map[string]any, keys ...string) any {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			return v
		}
	}
	return nil
}
