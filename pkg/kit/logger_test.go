package kit

import (
	"bytes"
	"context"
	"errors"
	"io"
	"iter"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	composeapi "github.com/docker/compose/v5/pkg/api"
	"github.com/moby/moby/api/types/jsonstream"
)

func TestLoggerFormatsPullBuildCompose(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := NewLogger(OutputStream(&buf), Color(false), Progress("plain"))
	log.Emit(LogEvent{Kind: KindPull, Op: "downloading", ID: "abc123", Parent: "alpine:latest", Current: 512, Total: 1024, Status: StatusWorking})
	log.Emit(LogEvent{Kind: KindPull, Op: "pulled", ID: "alpine:latest", Status: StatusDone})
	log.Emit(LogEvent{Kind: KindBuild, Op: "step", ID: "app", Detail: "Step 1/2 : FROM alpine", Status: StatusWorking})
	log.Emit(LogEvent{Kind: KindBuild, Op: "built", ID: "app", Detail: "sha256:deadbeef", Status: StatusDone})
	log.Emit(LogEvent{Kind: KindCompose, Op: "up", ID: "app", Status: StatusWorking})
	log.Emit(LogEvent{Kind: KindCompose, Op: "up", ID: "app", Status: StatusDone})
	got := buf.String()
	for _, want := range []string{
		"pull", "●", "abc123", "downloading", "50%",
		"✓", "alpine:latest", "pulled",
		"build", "step", "FROM alpine",
		"built", "sha256:deadbeef",
		"compose", "up",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "\x1b[") {
		t.Fatalf("plain output contains ansi: %q", got)
	}
}

func TestLoggerQuietSkipsWorking(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := NewLogger(OutputStream(&buf), Color(false), Progress("plain"), Quiet(true))
	log.Emit(LogEvent{Kind: KindPull, Op: "downloading", ID: "abc", Status: StatusWorking})
	log.Emit(LogEvent{Kind: KindPull, Op: "pulled", ID: "alpine", Status: StatusDone})
	got := buf.String()
	if strings.Contains(got, "downloading") {
		t.Fatalf("quiet leaked working: %s", got)
	}
	if !strings.Contains(got, "pulled") {
		t.Fatalf("quiet hid done: %s", got)
	}
}

func TestLoggerDrainsEngineJSONStreams(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := NewLogger(OutputStream(&buf), Color(false), Progress("plain"))
	pull := strings.Join([]string{
		`{"status":"Pulling from library/alpine","id":"latest"}`,
		`{"status":"Downloading","id":"abc123","progressDetail":{"current":512,"total":1024}}`,
		`{"status":"Pull complete","id":"abc123"}`,
		`{"status":"Status: Downloaded newer image for alpine:latest"}`,
	}, "\n")
	if err := log.drainJSON(context.Background(), strings.NewReader(pull), KindPull, "alpine:latest"); err != nil {
		t.Fatal(err)
	}
	build := strings.Join([]string{
		`{"stream":"Step 1/2 : FROM alpine\n"}`,
		`{"aux":{"ID":"sha256:deadbeef"}}`,
	}, "\n")
	if err := log.drainJSON(context.Background(), strings.NewReader(build), KindBuild, "app"); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{"pull", "pulling", "downloading", "50%", "pulled", "build", "step", "FROM alpine", "built", "sha256:deadbeef"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestLoggerComposeErrorAndDoneFlag(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := NewLogger(OutputStream(&buf), Color(false), Progress("plain"))
	log.On(composeapi.Resource{
		ID:     "service:imgproxy:1",
		Status: composeapi.Error,
		Text:   "Error response from daemon: Conflict. The container name \"/supabase-imgproxy\" is already in use",
	})
	log.Done("up", true)
	log.Done("down", false)
	got := buf.String()
	if strings.Contains(got, "error respo") {
		t.Fatalf("error message was truncated into the op column:\n%s", got)
	}
	if !strings.Contains(got, "already in use") || !strings.Contains(got, "service:imgproxy:1") {
		t.Fatalf("daemon error missing:\n%s", got)
	}
	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) != 3 {
		t.Fatalf("lines = %d\n%s", len(lines), got)
	}
	if !strings.Contains(lines[1], "✖") || !strings.Contains(lines[1], "up") {
		t.Fatalf("failed up marked success:\n%s", lines[1])
	}
	if !strings.Contains(lines[2], "✓") || !strings.Contains(lines[2], "down") {
		t.Fatalf("successful down marked failed:\n%s", lines[2])
	}
}

func TestLoggerComposeProgressAndLogs(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := NewLogger(OutputStream(&buf), Color(false), Progress("plain"))
	log.On(composeapi.Resource{ID: "api", Text: composeapi.StatusCreating, Status: composeapi.Working})
	log.On(composeapi.Resource{ID: "api", Text: composeapi.StatusCreated, Status: composeapi.Done})
	log.Log("api-1", "listening\n")
	log.Err("api-1", "warn\n")
	got := buf.String()
	for _, want := range []string{"compose", "creating", "created", "api-1", "│", "listening", "!", "warn"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestLoggerContainerLogLine(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := NewLogger(OutputStream(&buf), Color(false), Progress("plain"))
	log.Log("supabase-studio", "▲ Next.js 16.4.3\n")
	log.Err("supabase-db", "FATAL: password authentication failed\n")
	got := buf.String()
	if strings.Contains(got, "stdout") || strings.Contains(got, "stderr") || strings.Contains(got, "log ") {
		t.Fatalf("log row still uses the progress columns:\n%s", got)
	}
	if !strings.Contains(got, "supabase-studio") || !strings.Contains(got, "│ ▲ Next.js 16.4.3") {
		t.Fatalf("stdout row:\n%s", got)
	}
	if !strings.Contains(got, "supabase-db") || !strings.Contains(got, "! FATAL:") {
		t.Fatalf("stderr row:\n%s", got)
	}
}

func TestLoggerAttachedUpDropsTheLoadingRow(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := NewLogger(OutputStream(&buf), Color(false), Progress("tty"))
	log.spinEvery = -1
	log.Emit(LogEvent{Kind: KindCompose, Op: "up", ID: "supabase", Status: StatusWorking})
	if !strings.Contains(visibleScreen(buf.String()), "up") {
		t.Fatalf("startup row missing:\n%s", visibleScreen(buf.String()))
	}
	log.On(composeapi.Resource{ID: "Container supabase-edge", Text: "Waiting", Status: composeapi.Working})
	got := visibleScreen(buf.String())
	if fieldCount(got, "up") != 0 {
		t.Fatalf("project up row still on the loading screen:\n%s", got)
	}
	if !strings.Contains(got, "waiting") {
		t.Fatalf("container wait missing:\n%s", got)
	}
	log.Emit(LogEvent{Kind: KindCompose, Op: "up", ID: "supabase", Status: StatusDone})
	got = visibleScreen(buf.String())
	if fieldCount(got, "up") != 1 || !strings.Contains(got, "✓") {
		t.Fatalf("finished up line:\n%s", got)
	}
}

func fieldCount(screen, word string) int {
	n := 0
	for _, line := range strings.Split(screen, "\n") {
		for _, field := range strings.Fields(line) {
			if field == word {
				n++
			}
		}
	}
	return n
}

func TestLoggerHealthyClearsWaiting(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := NewLogger(OutputStream(&buf), Color(false), Progress("tty"))
	waiting := composeapi.Resource{ID: "Container supabase-db", Text: "Waiting", Status: composeapi.Working}
	healthy := composeapi.Resource{ID: "Container supabase-db", Text: "Healthy", Status: composeapi.Done}
	log.On(waiting)
	log.On(healthy)
	log.On(waiting)
	log.On(healthy)
	got := visibleScreen(buf.String())
	if strings.Contains(got, "waiting") {
		t.Fatalf("waiting row stuck after the container was healthy:\n%s", got)
	}
	if strings.Count(buf.String(), "healthy") != 1 {
		t.Fatalf("healthy printed more than once:\n%s", buf.String())
	}
}

func TestVisualRowsCountsWrappedLines(t *testing.T) {
	t.Parallel()
	if visualRows("hello", 0) != 1 || visualRows("hello", 20) != 1 {
		t.Fatal("short line")
	}
	if visualRows(strings.Repeat("a", 45), 20) != 3 {
		t.Fatal("wrapped line")
	}
	colored := "\x1b[1m" + strings.Repeat("a", 45) + "\x1b[0m"
	if visualRows(colored, 20) != 3 || visibleCols(colored) != 45 {
		t.Fatal("color sequences counted as columns")
	}
}

func TestLoggerLongLogClearsWrappedRows(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := NewLogger(OutputStream(&buf), Color(false), Progress("tty"))
	log.cols = 20
	log.On(composeapi.Resource{ID: "Container supabase-db", Text: "Waiting", Status: composeapi.Working})
	msg := strings.Repeat("word ", 12) + "tail"
	log.Log("supabase-kong", msg)
	got := visibleScreen(buf.String())
	if strings.Contains(got, "tailwaiting") || strings.Contains(got, "tail waiting") {
		t.Fatalf("spinner tail stuck to the log line:\n%s", got)
	}
	if !strings.Contains(got, "tail") || !strings.Contains(got, "waiting") {
		t.Fatalf("log or spinner missing:\n%s", got)
	}
}

func TestLoggerComposeStatusIsNotRepeated(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := NewLogger(OutputStream(&buf), Color(false), Progress("plain"))
	created := composeapi.Resource{ID: "Container supabase-imgproxy", Text: "Created", Status: composeapi.Done}
	log.On(created)
	log.On(created)
	healthy := composeapi.Resource{ID: "Container supabase-db", Text: "Healthy", Status: composeapi.Done}
	log.On(healthy)
	log.On(healthy)
	log.On(composeapi.Resource{ID: "Container supabase-db", Text: "Started", Status: composeapi.Done})
	log.On(composeapi.Resource{
		ID:      "Volume vector-config",
		Text:    "Resource is still in use",
		Details: "Resource is still in use",
		Status:  composeapi.Warning,
	})
	log.Log("api", "\n")
	log.Log("api", "listening\n")

	got := buf.String()
	if strings.Contains(got, "Created") || strings.Contains(got, "Healthy") || strings.Contains(got, "Started") {
		t.Fatalf("status repeated beside the operation:\n%s", got)
	}
	if strings.Count(got, "created") != 1 || strings.Count(got, "healthy") != 1 || strings.Count(got, "started") != 1 {
		t.Fatalf("done event repeated:\n%s", got)
	}
	if !strings.Contains(got, "Resource is still in use") {
		t.Fatalf("long status dropped:\n%s", got)
	}
	if strings.Count(got, "│") != 1 || !strings.Contains(got, "listening") {
		t.Fatalf("blank log line kept:\n%s", got)
	}
}

func TestLogSetterWiresComposeAndEngine(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := NewLogger(OutputStream(&buf), Color(false), Progress("plain"))
	p := NewCompose("app", AddService(Name("api"), Image("alpine:latest")))
	api := &fakeComposer{}
	ctx := context.Background()
	if err := p.Up(ctx, api, Log(log)); err != nil {
		t.Fatal(err)
	}
	if api.upOpts.Start.Attach != log {
		t.Fatal("up should attach logger as log consumer")
	}
	if !api.upOpts.Create.QuietPull {
		t.Fatal("log should silence stock compose pull")
	}
	if err := p.Build(ctx, api, Log(log)); err != nil {
		t.Fatal(err)
	}
	if !api.built.Quiet || api.built.Out != io.Discard {
		t.Fatalf("build stock not silenced: %+v", api.built)
	}
	if err := p.Pull(ctx, api, Log(log)); err != nil {
		t.Fatal(err)
	}
	if !api.pulled.Quiet {
		t.Fatal("pull stock not silenced")
	}
	got := buf.String()
	for _, want := range []string{"compose", "up", "build", "pull"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}

	pullJSON := `{"status":"Downloading","id":"layer","progressDetail":{"current":1,"total":2}}` + "\n"
	eng := &fakeEngine{pullResp: jsonPull{ReadCloser: io.NopCloser(strings.NewReader(pullJSON))}}
	if err := PullImage(ctx, eng, "alpine:latest", Log(log)); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "downloading") {
		t.Fatalf("engine pull not logged:\n%s", buf.String())
	}
	buildJSON := `{"stream":"Step 1/1 : FROM alpine\n"}` + "\n"
	eng.buildBody = io.NopCloser(strings.NewReader(buildJSON))
	if err := BuildImage(ctx, eng, NewDockerfile().From("alpine:latest"), Image("app:dev"), Log(log)); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "FROM alpine") {
		t.Fatalf("engine build not logged:\n%s", buf.String())
	}
}

type jsonPull struct{ io.ReadCloser }

func (jsonPull) JSONMessages(context.Context) iter.Seq2[jsonstream.Message, error] {
	return func(func(jsonstream.Message, error) bool) {}
}

func (jsonPull) Wait(context.Context) error { return nil }

var _ ImagePullResponse = jsonPull{}

func TestLoggerTTYPullUpdatesInPlace(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := NewLogger(OutputStream(&buf), Color(false), Progress("tty"))
	log.spinEvery = -1

	log.Emit(LogEvent{Kind: KindPull, Op: "pulling", ID: "alpine:latest", Status: StatusWorking})
	first := buf.String()
	if strings.Contains(first, "\x1b[1A") {
		t.Fatalf("first loading row moved the cursor: %q", first)
	}
	if !strings.Contains(visibleScreen(first), "⠋") || !strings.Contains(visibleScreen(first), "pulling") {
		t.Fatalf("loading row:\n%s", visibleScreen(first))
	}

	log.Emit(LogEvent{Kind: KindPull, Op: "downloading", ID: "abc123", Parent: "alpine:latest", Current: 512, Total: 1024, Status: StatusWorking})
	log.Emit(LogEvent{Kind: KindPull, Op: "downloading", ID: "abc123", Parent: "alpine:latest", Current: 1024, Total: 1024, Status: StatusWorking})
	log.advanceSpinner()
	got := visibleScreen(buf.String())
	if strings.Contains(got, "50%") {
		t.Fatalf("stale progress stayed on screen:\n%s", got)
	}
	if !strings.Contains(got, "100%") || !strings.Contains(got, "⠙") || !strings.Contains(got, "pulling") {
		t.Fatalf("live pull rows:\n%s\nraw:\n%q", got, buf.String())
	}
	if !strings.Contains(buf.String(), "\x1b[2A") {
		t.Fatalf("layer update did not redraw in place: %q", buf.String())
	}

	log.Emit(LogEvent{Kind: KindLog, Op: "stdout", ID: "api", Detail: "listening", Status: StatusInfo})
	got = visibleScreen(buf.String())
	if !strings.Contains(got, "listening") || !strings.Contains(got, "100%") {
		t.Fatalf("log line ate the loading rows:\n%s", got)
	}

	log.Emit(LogEvent{Kind: KindPull, Op: "pulled", ID: "abc123", Parent: "alpine:latest", Status: StatusDone})
	log.Emit(LogEvent{Kind: KindPull, Op: "pulled", ID: "alpine:latest", Status: StatusDone})
	got = visibleScreen(buf.String())
	if strings.Contains(got, "⠋") || strings.Contains(got, "⠙") || strings.Contains(got, "100%") {
		t.Fatalf("finished pull left a live row:\n%s", got)
	}
	if strings.Count(got, "✓") < 2 || !strings.Contains(got, "listening") {
		t.Fatalf("transcript:\n%s", got)
	}
}

func TestLoggerTTYBuildKeepsFinishedSteps(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := NewLogger(OutputStream(&buf), Color(false), Progress("tty"))
	log.spinEvery = -1
	build := strings.Join([]string{
		`{"stream":"Step 1/2 : FROM alpine\n"}`,
		`{"stream":" ---> abc\n"}`,
		`{"stream":"Step 2/2 : RUN echo hi\n"}`,
		`{"aux":{"ID":"sha256:deadbeef"}}`,
	}, "\n")
	emitPhase(log, KindBuild, "building", "app:dev", StatusWorking, "")
	if err := log.drainJSON(context.Background(), strings.NewReader(build), KindBuild, "app:dev"); err != nil {
		t.Fatal(err)
	}
	finishPhase(log, KindBuild, "built", "app:dev")
	got := visibleScreen(buf.String())
	for _, want := range []string{"Step 1/2 : FROM alpine", "Step 2/2 : RUN echo hi", "---> abc", "sha256:deadbeef", "✓"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s\nraw:\n%q", want, got, buf.String())
		}
	}
	if strings.Contains(got, "⠋") || strings.Contains(got, "building") {
		t.Fatalf("build loading row still live:\n%s", got)
	}
}

func TestLoggerTTYSpinnerTicks(t *testing.T) {
	t.Parallel()
	var buf syncBuf
	log := NewLogger(OutputStream(&buf), Color(false), Progress("tty"))
	log.spinEvery = 15 * time.Millisecond
	log.Emit(LogEvent{Kind: KindPull, Op: "pulling", ID: "alpine:latest", Status: StatusWorking})

	deadline := time.Now().Add(time.Second)
	var moved bool
	for time.Now().Before(deadline) {
		if strings.Contains(buf.String(), "⠙") {
			moved = true
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if !moved {
		t.Fatalf("spinner did not advance:\n%s", buf.String())
	}

	log.Emit(LogEvent{Kind: KindPull, Op: "pulled", ID: "alpine:latest", Status: StatusDone})
	deadline = time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		log.mu.Lock()
		on := log.spinning
		log.mu.Unlock()
		if !on {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("spinner kept running after the pull finished")
}

func TestPullAndBuildShowLoadingRow(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := NewLogger(OutputStream(&buf), Color(false), Progress("tty"))
	log.spinEvery = -1

	pullDuring := make(chan string, 1)
	pull := strings.Join([]string{
		`{"status":"Pulling from library/alpine","id":"latest"}`,
		`{"status":"Downloading","id":"abc123","progressDetail":{"current":512,"total":1024}}`,
		`{"status":"Downloading","id":"abc123","progressDetail":{"current":1024,"total":1024}}`,
		`{"status":"Pull complete","id":"abc123"}`,
		`{"status":"Status: Downloaded newer image for alpine:latest"}`,
	}, "\n")
	eng := &fakeEngine{
		onPull:   func() { pullDuring <- visibleScreen(buf.String()) },
		pullResp: jsonPull{ReadCloser: io.NopCloser(strings.NewReader(pull))},
	}
	if err := PullImage(context.Background(), eng, "alpine:latest", Log(log)); err != nil {
		t.Fatal(err)
	}
	gotDuring := <-pullDuring
	if !strings.Contains(gotDuring, "⠋") || !strings.Contains(gotDuring, "pulling") {
		t.Fatalf("pull loading row missing while the engine runs:\n%s", gotDuring)
	}
	got := visibleScreen(buf.String())
	if strings.Contains(got, "⠋") || strings.Contains(got, "50%") {
		t.Fatalf("pull left a live or stale row:\n%s", got)
	}
	if !strings.Contains(got, "pulled") || !strings.Contains(got, "✓") {
		t.Fatalf("pull transcript:\n%s\nraw:\n%q", got, buf.String())
	}

	buildDuring := make(chan string, 1)
	eng.onBuild = func() { buildDuring <- visibleScreen(buf.String()) }
	eng.buildBody = io.NopCloser(strings.NewReader(`{"stream":"Step 1/1 : FROM alpine\n"}` + "\n" + `{"aux":{"ID":"sha256:deadbeef"}}` + "\n"))
	if err := BuildImage(context.Background(), eng, NewDockerfile().From("alpine:latest"), Image("app:dev"), Log(log)); err != nil {
		t.Fatal(err)
	}
	gotDuring = <-buildDuring
	if !strings.Contains(gotDuring, "building") || !strings.Contains(gotDuring, "⠋") {
		t.Fatalf("build loading row missing while the engine runs:\n%s", gotDuring)
	}
	got = visibleScreen(buf.String())
	if !strings.Contains(got, "FROM alpine") || !strings.Contains(got, "sha256:deadbeef") {
		t.Fatalf("build transcript:\n%s", got)
	}
	if strings.Contains(got, "⠋") {
		t.Fatalf("build left a spinner:\n%s", got)
	}
}

func TestPullImageLoadingRowOnError(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := NewLogger(OutputStream(&buf), Color(false), Progress("tty"))
	log.spinEvery = -1
	eng := &fakeEngine{pullErr: errors.New("denied")}
	err := PullImage(context.Background(), eng, "alpine:latest", Log(log))
	if err == nil {
		t.Fatal("expected pull error")
	}
	got := visibleScreen(buf.String())
	if !strings.Contains(got, "✖") || !strings.Contains(got, "denied") {
		t.Fatalf("error row:\n%s", got)
	}
	if strings.Contains(got, "⠋") {
		t.Fatalf("spinner left after pull error:\n%s", got)
	}
}

func TestLoggerTTYComposePullAndBuild(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := NewLogger(OutputStream(&buf), Color(false), Progress("tty"))
	log.spinEvery = -1
	err := logAround(context.Background(), log, KindPull, "pull", "app", func(context.Context) error {
		got := visibleScreen(buf.String())
		if !strings.Contains(got, "⠋") || !strings.Contains(got, "pull") {
			t.Fatalf("compose pull loading row:\n%s", got)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = logAround(context.Background(), log, KindBuild, "build", "app", func(context.Context) error {
		got := visibleScreen(buf.String())
		if !strings.Contains(got, "⠋") || !strings.Contains(got, "build") {
			t.Fatalf("compose build loading row:\n%s", got)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	got := visibleScreen(buf.String())
	if strings.Contains(got, "⠋") || strings.Count(got, "✓") < 2 {
		t.Fatalf("compose transcript:\n%s", got)
	}
}

type syncBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuf) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuf) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

func visibleScreen(s string) string {
	lines := []string{""}
	row := 0
	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && ((s[j] >= '0' && s[j] <= '9') || s[j] == ';') {
				j++
			}
			if j >= len(s) {
				break
			}
			n := 1
			if arg := s[i+2 : j]; arg != "" {
				if v, err := strconv.Atoi(arg); err == nil {
					n = v
				}
			}
			switch s[j] {
			case 'A':
				row -= n
				if row < 0 {
					row = 0
				}
			case 'K':
				for len(lines) <= row {
					lines = append(lines, "")
				}
				lines[row] = ""
			}
			i = j + 1
			continue
		}
		switch s[i] {
		case '\r':
			i++
		case '\n':
			row++
			for len(lines) <= row {
				lines = append(lines, "")
			}
			i++
		default:
			for len(lines) <= row {
				lines = append(lines, "")
			}
			lines[row] += s[i : i+1]
			i++
		}
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}

func TestNilLoggerDrainsPushErrors(t *testing.T) {
	t.Parallel()
	body := `{"errorDetail":{"message":"unauthorized: authentication required"},"error":"unauthorized: authentication required"}` + "\n"
	var log *Logger
	err := log.drainJSON(context.Background(), strings.NewReader(body), KindPush, "127.0.0.1:5000/app")
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
	if !strings.Contains(err.Error(), "authentication required") {
		t.Fatalf("err = %v", err)
	}
}
