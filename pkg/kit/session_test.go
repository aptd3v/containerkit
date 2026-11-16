package kit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	composeapi "github.com/aptd3v/containerkit/internal/compose/api/types"
)

func TestErrorClassNotFound(t *testing.T) {
	t.Parallel()
	err := cfgErr("engine", "alpine:latest", "pullImage", ErrNotFound)
	if got := ErrorClass(err); got != "notFound" {
		t.Fatalf("class = %q", got)
	}
	rep := errorReport(err)
	if rep.Receiver != "engine" {
		t.Fatalf("receiver = %q", rep.Receiver)
	}
	if rep.Path != "alpine:latest" {
		t.Fatalf("path = %q", rep.Path)
	}
	if strings.Contains(rep.Message, "supersecret") {
		t.Fatal(rep.Message)
	}
}

func TestErrorReportExitAndSecret(t *testing.T) {
	t.Parallel()
	err := cfgErr("container", "wait", "Wait", &ExitError{StatusCode: 2})
	rep := errorReport(err)
	if rep.Class != "internal" {
		t.Fatalf("class = %q", rep.Class)
	}
	if rep.ExitCode == nil || *rep.ExitCode != 2 {
		t.Fatalf("exit = %v", rep.ExitCode)
	}
	secret := secretErr("engine", "login", "Login", errors.New("supersecret"))
	rep = errorReport(secret)
	if strings.Contains(rep.Message, "supersecret") {
		t.Fatalf("secret leaked: %s", rep.Message)
	}
}

func TestSessionConnectCarriesLogger(t *testing.T) {
	t.Parallel()
	var saw LogEvent
	log := NewFrameLogger(func(e LogEvent) { saw = e })
	s := &Session{log: log}
	act, err := collectComposerAct(s.connectOpts(nil))
	if err != nil {
		t.Fatal(err)
	}
	if act.log != log {
		t.Fatal("session logger was not attached to the composer")
	}
	act.log.On(composeapi.Resource{ID: "Container supabase-imgproxy", Text: "recreated", Status: composeapi.Done})
	if saw.ID != "Container supabase-imgproxy" || saw.Status != StatusDone {
		t.Fatalf("container event = %#v", saw)
	}
	frame := eventFrame(saw)
	display, _ := frame["display"].(string)
	if display != eventDisplay(saw) {
		t.Fatalf("display = %q", display)
	}
	if !strings.Contains(display, "compose") || !strings.Contains(display, "Container s…e-imgproxy") || !strings.Contains(display, "recreated") {
		t.Fatalf("column line = %q", display)
	}
	if strings.Contains(display, "supabase-imgproxy") {
		t.Fatalf("id was not clipped to the logger column: %q", display)
	}
}

func TestEventDisplayMatchesLoggerAndColor(t *testing.T) {
	t.Parallel()
	e := LogEvent{Kind: KindCompose, Op: "up", ID: "supabase", Status: StatusDone}
	if eventDisplay(e) != formatEvent(e, eventColor(), false, statusGlyph(e.Status)) {
		t.Fatalf("display = %q", eventDisplay(e))
	}
	colored := formatEvent(e, true, false, statusGlyph(e.Status))
	if !strings.Contains(colored, "\x1b[") {
		t.Fatalf("colored row = %q", colored)
	}
	if eventColorFrom("", "1") != true || eventColorFrom("1", "1") != false || eventColorFrom("", "") != false {
		t.Fatal("color flag")
	}
}

func TestServeSessionInvalidPlan(t *testing.T) {
	t.Parallel()
	var out, errw bytes.Buffer
	in := strings.NewReader("{\"id\":1,\"type\":\"exec\",\"plan\":{\"actions\":[{\"op\":\"\"}]}}\n")
	code := ServeSession(context.Background(), in, &out, &errw)
	if code != 0 {
		t.Fatalf("code = %d, err = %s", code, errw.String())
	}
	var frame sessionError
	if err := json.Unmarshal(out.Bytes(), &frame); err != nil {
		t.Fatal(err)
	}
	if frame.Type != "error" || frame.Class != "invalidArgument" {
		t.Fatalf("frame = %+v", frame)
	}
	if string(frame.ID) != "1" {
		t.Fatalf("id = %s", frame.ID)
	}
}

func TestServeLogAndResult(t *testing.T) {
	t.Parallel()
	var out, errw bytes.Buffer
	w := &frameWriter{enc: json.NewEncoder(&out)}
	log := NewFrameLogger(func(e LogEvent) {
		if err := w.write(eventFrame(e)); err != nil {
			t.Error(err)
		}
	})
	in := strings.NewReader("{\"id\":7,\"type\":\"exec\",\"plan\":{}}\n")
	code := serve(context.Background(), in, &errw, w, func(ctx context.Context, _ []byte) ([]byte, error) {
		loggerFrom(log.Context(ctx)).Log("api", "hello")
		loggerFrom(log.Context(ctx)).Status("api", "started")
		return []byte(`{"ok":true}`), nil
	})
	if code != 0 {
		t.Fatalf("code = %d, err = %s", code, errw.String())
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("lines = %q", lines)
	}
	var logFrame map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &logFrame); err != nil {
		t.Fatal(err)
	}
	if logFrame["kind"] != "log" || logFrame["source"] != "api" || logFrame["stream"] != "stdout" || logFrame["line"] != "hello" {
		t.Fatalf("log = %#v", logFrame)
	}
	wantLog := eventDisplay(LogEvent{Kind: KindLog, Op: "stdout", ID: "api", Detail: "hello", Status: StatusInfo})
	if logFrame["group"] != "log" || logFrame["status"] != "info" || logFrame["display"] != wantLog || !strings.Contains(wantLog, "log") || !strings.Contains(wantLog, "stdout") || !strings.Contains(wantLog, "hello") {
		t.Fatalf("log display = %#v, want %q", logFrame, wantLog)
	}
	var status map[string]any
	if err := json.Unmarshal([]byte(lines[1]), &status); err != nil {
		t.Fatal(err)
	}
	if status["kind"] != "status" || status["line"] != "started" {
		t.Fatalf("status = %#v", status)
	}
	wantStatus := eventDisplay(LogEvent{Kind: KindCompose, Op: "status", ID: "api", Detail: "started", Status: StatusInfo})
	if status["group"] != "compose" || status["status"] != "info" || status["display"] != wantStatus || !strings.Contains(wantStatus, "compose") || !strings.Contains(wantStatus, "started") {
		t.Fatalf("status display = %#v, want %q", status, wantStatus)
	}
	var result sessionResult
	if err := json.Unmarshal([]byte(lines[2]), &result); err != nil {
		t.Fatal(err)
	}
	if result.Type != "result" || !result.OK || string(result.ID) != "7" {
		t.Fatalf("result = %+v", result)
	}
}
