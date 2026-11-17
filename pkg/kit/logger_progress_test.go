package kit

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	composeapi "github.com/aptd3v/containerkit/internal/compose/api/types"
	composetypes "github.com/aptd3v/containerkit/internal/compose/types"
)

func TestLoggerJSONModeEmitsOneRecordPerEvent(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := NewLogger(OutputStream(&buf), Progress("json"))
	if log.tty {
		t.Fatal("json mode used a live region")
	}
	if tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0); err == nil {
		defer tty.Close()
		term := NewLogger(OutputStream(tty), Progress("json"))
		if term.tty {
			t.Fatal("json mode followed the terminal")
		}
	}
	event := LogEvent{Kind: KindPull, Op: "downloading", ID: "abc", Text: "Downloading", Current: 1, Total: 2, Status: StatusWorking}
	log.Emit(event)
	log.Emit(LogEvent{Kind: KindLog, Op: "stdout", ID: "api", Detail: "hi", Status: StatusInfo})
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("lines = %q", lines)
	}
	var frame map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &frame); err != nil {
		t.Fatal(err)
	}
	want := eventFrame(event)
	for _, key := range []string{"type", "kind", "group", "status", "source", "stream", "line", "text", "current", "total", "display"} {
		got, gok := frame[key]
		exp, eok := want[key]
		if gok != eok || (eok && got != jsonNumber(exp)) {
			t.Fatalf("field %s = %#v, want %#v", key, got, exp)
		}
	}
	if strings.Contains(buf.String(), "\x1b[") {
		t.Fatalf("json row contains a live-region sequence: %s", buf.String())
	}
	var quiet bytes.Buffer
	q := NewLogger(OutputStream(&quiet), Progress("json"), Quiet(true))
	q.Emit(event)
	q.Emit(LogEvent{Kind: KindPull, Op: "pulled", ID: "abc", Status: StatusDone})
	if strings.Contains(quiet.String(), "downloading") || !strings.Contains(quiet.String(), "pulled") {
		t.Fatalf("quiet json = %s", quiet.String())
	}
}

func jsonNumber(v any) any {
	switch n := v.(type) {
	case int64:
		return float64(n)
	case int:
		return float64(n)
	default:
		return v
	}
}

type rowComposer struct {
	fakeComposer
	bridge *logBridge
}

func (c *rowComposer) Up(ctx context.Context, project *composetypes.Project, options composeapi.UpOptions) error {
	c.bridge.Start(ctx, "up")
	c.bridge.On(composeapi.Resource{ID: "Container api", Text: "Started", Status: composeapi.Done})
	c.bridge.Done("up", false)
	return c.fakeComposer.Up(ctx, project, options)
}

func TestLogOnUpReceivesContainerRow(t *testing.T) {
	t.Parallel()
	var events []LogEvent
	log := NewFrameLogger(func(e LogEvent) { events = append(events, e) })
	api := &rowComposer{bridge: newLogBridge(nil)}
	p := NewCompose("app").WithService(NewService(Name("api"), Image("alpine:latest")))
	if err := p.Up(context.Background(), api, Log(log)); err != nil {
		t.Fatal(err)
	}
	var container, bareUp int
	for _, e := range events {
		if e.ID == "Container api" {
			container++
		}
		if e.Op == "up" && e.ID == "" {
			bareUp++
		}
	}
	if container != 1 {
		t.Fatalf("container events = %d in %#v", container, events)
	}
	if bareUp != 0 {
		t.Fatalf("bridge forwarded a second up line: %#v", events)
	}
}

func TestConnectLoggerReceivesComposeStart(t *testing.T) {
	t.Parallel()
	var events []LogEvent
	log := NewFrameLogger(func(e LogEvent) { events = append(events, e) })
	bridge := newLogBridge(log)
	bridge.Start(context.Background(), "up")
	bridge.On(composeapi.Resource{ID: "Container api", Text: "Started", Status: composeapi.Done})
	bridge.Done("up", false)
	var working, container, done int
	for _, e := range events {
		if e.Op == "up" && e.Status == StatusWorking && e.ID == "" {
			working++
		}
		if e.ID == "Container api" {
			container++
		}
		if e.Op == "up" && e.Status == StatusDone && e.ID == "" {
			done++
		}
	}
	if working != 1 || container != 1 || done != 1 {
		t.Fatalf("events = %#v", events)
	}
}

type scriptedLoad struct {
	*fakeEngine
	body ImageLoadResult
}

func (s *scriptedLoad) ImageLoad(context.Context, io.Reader, ...ImageLoadOption) (ImageLoadResult, error) {
	return s.body, nil
}

func TestLoadImageDrainsWhenLogged(t *testing.T) {
	t.Parallel()
	raw := `{"status":"Loading layer","id":"abc"}` + "\n"
	plain, err := LoadImage(context.Background(), &scriptedLoad{
		fakeEngine: &fakeEngine{},
		body:       io.NopCloser(strings.NewReader(raw)),
	}, strings.NewReader("tar"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(plain)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != raw {
		t.Fatalf("body = %q", got)
	}

	var events []LogEvent
	log := NewFrameLogger(func(e LogEvent) { events = append(events, e) })
	drained, err := LoadImage(context.Background(), &scriptedLoad{
		fakeEngine: &fakeEngine{},
		body:       io.NopCloser(strings.NewReader(raw)),
	}, strings.NewReader("tar"), Log(log))
	if err != nil {
		t.Fatal(err)
	}
	left, err := io.ReadAll(drained)
	if err != nil {
		t.Fatal(err)
	}
	if len(left) != 0 {
		t.Fatalf("logged load returned %q", left)
	}
	var saw bool
	for _, e := range events {
		if e.Kind == KindEngine && e.ID == "abc" {
			saw = true
		}
	}
	if !saw {
		t.Fatalf("events = %#v", events)
	}
}

type scriptedLogs struct {
	*fakeEngine
	body ContainerLogsResult
}

func (s *scriptedLogs) ContainerLogs(context.Context, string, ContainerLogsOptions) (ContainerLogsResult, error) {
	return s.body, nil
}

func TestContainerLogsEmitsAndKeepsFrames(t *testing.T) {
	t.Parallel()
	raw := append(stdFrame(1, "out\n"), stdFrame(2, "err\n")...)
	var events []LogEvent
	log := NewFrameLogger(func(e LogEvent) { events = append(events, e) })
	rc, err := NewContainer().Logs(context.Background(), &scriptedLogs{
		fakeEngine: &fakeEngine{},
		body:       io.NopCloser(bytes.NewReader(raw)),
	}, "box", Log(log))
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatal(err)
	}
	if err := rc.Close(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, raw) {
		t.Fatalf("bytes = %q", got)
	}
	var stdout, stderr bool
	for _, e := range events {
		if e.ID == "box" && e.Op == "stdout" && e.Detail == "out" {
			stdout = true
		}
		if e.ID == "box" && e.Op == "stderr" && e.Detail == "err" {
			stderr = true
		}
	}
	if !stdout || !stderr {
		t.Fatalf("events = %#v", events)
	}

	tty := io.NopCloser(strings.NewReader("hello\n"))
	rc, err = NewContainer().Logs(log.Context(context.Background()), &scriptedLogs{
		fakeEngine: &fakeEngine{},
		body:       tty,
	}, "box")
	if err != nil {
		t.Fatal(err)
	}
	if err := consumeContainerLogs(rc); err != nil {
		t.Fatal(err)
	}
	var hello bool
	for _, e := range events {
		if e.Detail == "hello" && e.Op == "stdout" {
			hello = true
		}
	}
	if !hello {
		t.Fatalf("tty line missing: %#v", events)
	}
}

func stdFrame(stream byte, payload string) []byte {
	b := make([]byte, 8+len(payload))
	b[0] = stream
	binary.BigEndian.PutUint32(b[4:], uint32(len(payload)))
	copy(b[8:], payload)
	return b
}

type fireComposer struct {
	fakeComposer
	called bool
}

func (f *fireComposer) Logs(context.Context, string, composeapi.LogConsumer, composeapi.LogOptions) error {
	f.called = true
	return nil
}

type callingComposer struct {
	fakeComposer
	consumer composeapi.LogConsumer
}

func (c *callingComposer) Logs(_ context.Context, _ string, consumer composeapi.LogConsumer, _ composeapi.LogOptions) error {
	c.consumer = consumer
	consumer.Log("api", "hello")
	return nil
}

type countConsumer struct{ n int }

func (c *countConsumer) Log(string, string)    { c.n++ }
func (c *countConsumer) Err(string, string)    {}
func (c *countConsumer) Status(string, string) {}

func TestComposeLogsUsesContextLogger(t *testing.T) {
	t.Parallel()
	p := NewCompose("app")
	fire := &fireComposer{}
	if err := p.Logs(context.Background(), fire, nil); err == nil || fire.called {
		t.Fatalf("expected a missing consumer, err=%v called=%v", err, fire.called)
	}
	var events []LogEvent
	log := NewFrameLogger(func(e LogEvent) { events = append(events, e) })
	api := &callingComposer{}
	if err := p.Logs(log.Context(context.Background()), api, nil); err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].ID != "api" || events[0].Detail != "hello" {
		t.Fatalf("events = %#v", events)
	}
	count := &countConsumer{}
	events = nil
	if err := p.Logs(log.Context(context.Background()), api, count); err != nil {
		t.Fatal(err)
	}
	if count.n != 1 || len(events) != 1 {
		t.Fatalf("tee count = %d events = %#v", count.n, events)
	}
	events = nil
	if err := p.Logs(context.Background(), api, log, Log(log)); err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("logger passed twice delivered %d events", len(events))
	}
}

func TestPlanProgress(t *testing.T) {
	t.Parallel()
	if _, err := progressMode("verbose"); err == nil {
		t.Fatal("expected invalid progress")
	}
	if err := ImportPlan([]byte(`{"progress":"nope"}`)).Validate(); err == nil {
		t.Fatal("validate accepted nope")
	}
	var buf bytes.Buffer
	ctx, err := attachPlanProgress(context.Background(), "", &buf, false)
	if err != nil {
		t.Fatal(err)
	}
	if LoggerFrom(ctx) != nil {
		t.Fatal("empty progress logged")
	}
	ctx, err = attachPlanProgress(context.Background(), "off", &buf, true)
	if err != nil {
		t.Fatal(err)
	}
	if LoggerFrom(ctx) != nil {
		t.Fatal("off logged")
	}
	ctx, err = attachPlanProgress(context.Background(), "", &buf, true)
	if err != nil {
		t.Fatal(err)
	}
	if LoggerFrom(ctx) == nil {
		t.Fatal("exec default did not log")
	}
	buf.Reset()
	ctx, err = attachPlanProgress(context.Background(), "json", &buf, false)
	if err != nil {
		t.Fatal(err)
	}
	LoggerFrom(ctx).Emit(LogEvent{Kind: KindPull, Op: "pulled", ID: "alpine", Status: StatusDone})
	if !strings.Contains(buf.String(), `"group":"pull"`) {
		t.Fatalf("json progress = %s", buf.String())
	}
}
