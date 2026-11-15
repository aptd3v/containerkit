package kit

import (
	"context"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	composeapi "github.com/docker/compose/v5/pkg/api"
	"golang.org/x/term"
)

// LogKind names the source of a [LogEvent].
type LogKind string

const (
	KindPull    LogKind = "pull"
	KindPush    LogKind = "push"
	KindBuild   LogKind = "build"
	KindCompose LogKind = "compose"
	KindEngine  LogKind = "engine"
	KindLog     LogKind = "log"
)

// LogStatus is the lifecycle of a [LogEvent].
type LogStatus string

const (
	StatusWorking LogStatus = "working"
	StatusDone    LogStatus = "done"
	StatusWarn    LogStatus = "warn"
	StatusError   LogStatus = "error"
	StatusInfo    LogStatus = "info"
)

// LogEvent is one unified progress or log line from Engine or Compose.
type LogEvent struct {
	Time    time.Time
	Kind    LogKind
	Op      string
	ID      string
	Parent  string
	Text    string
	Detail  string
	Current int64
	Total   int64
	Status  LogStatus
}

// Logger renders Engine and Compose progress with one vocabulary.
// It does not use Docker's stock JSON-message printer or Compose TTY writer:
// JSON streams and Compose events are parsed and formatted here.
//
// In tty mode, pull and build rows stay on a live region at the bottom of the
// terminal. Working rows show a spinner, byte totals update in place, and a
// finished row scrolls into the transcript above that region.
type Logger struct {
	mu         sync.Mutex
	out        io.Writer
	color      field[bool]
	timestamps bool
	quiet      bool
	mode       string
	tty        bool
	live       []string
	rows       map[string]LogEvent
	lastDone   map[string]string
	pending    []string // log lines waiting for one paint of the live region
	drawn      int      // visual rows occupied by the live region
	cols       int      // terminal width override; 0 measures, negative disables wrap
	realTTY    bool
	spin       int
	spinEvery  time.Duration // zero uses the default; negative disables the ticker
	spinning   bool
	frames     bool
	sink       func(LogEvent)
}

// NewLogger constructs a stdout logger. With no options it writes to stderr,
// auto-detects a TTY, and colors when the destination is a terminal.
func NewLogger(opts ...LoggerOption) *Logger {
	l := &Logger{
		out:      os.Stderr,
		mode:     "auto",
		rows:     map[string]LogEvent{},
		lastDone: map[string]string{},
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		_ = opt.applyLogger(l)
	}
	l.resolve()
	return l
}

func (l *Logger) resolve() {
	if l.out == nil {
		l.out = os.Stderr
	}
	switch l.mode {
	case "tty":
		l.tty = true
	case "plain", "quiet":
		l.tty = false
	default:
		l.tty = isTTY(l.out)
	}
	l.realTTY = isTTY(l.out)
	if l.mode == "quiet" {
		l.quiet = true
	}
}

func (l *Logger) colorOn() bool {
	if l.color.ok() {
		return l.color.val
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return l.tty
}

type loggerContextKey struct{}

// Context attaches this logger to ctx for Engine and Compose actions that
// honor it.
func (l *Logger) Context(ctx context.Context) context.Context {
	if l == nil {
		return ctx
	}
	return context.WithValue(ctx, loggerContextKey{}, l)
}

// Emit renders one event. Nil loggers and quiet working events are skipped.
func (l *Logger) Emit(e LogEvent) {
	if l == nil {
		return
	}
	if e.Status == "" {
		e.Status = StatusInfo
	}
	if l.quiet && e.Status == StatusWorking {
		return
	}
	if e.Time.IsZero() {
		e.Time = time.Now()
	}
	e.Op = oneLine(e.Op)
	e.Text = oneLine(e.Text)
	e.Detail = oneLine(e.Detail)
	if e.Kind == KindLog && strings.TrimSpace(e.Detail) == "" {
		return
	}
	if l.frames {
		if l.sink != nil {
			l.sink(e)
		}
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if e.Kind == KindCompose && e.Status == StatusDone && l.repeatDone(e) {
		// A later waiter reports the same healthy container again. The line
		// stays in the transcript once, but the live "waiting" row has to go.
		l.clearLiveLocked(eventKey(e))
		return
	}
	if !l.tty {
		_, _ = io.WriteString(l.out, l.format(e)+"\n")
		return
	}
	// A chatty container (Kong writes a line per worker) would otherwise
	// repaint the live region on every line and the terminal falls behind.
	// Hold those lines for the spinner tick, at most one frame later.
	if e.Kind == KindLog && l.realTTY && len(l.live) > 0 && l.spinEvery >= 0 {
		l.pending = append(l.pending, l.format(e))
		if l.hidingProjectUp(e) {
			l.retireProjectUpLocked()
		}
		return
	}
	if l.hidingProjectUp(e) {
		l.retireProjectUpLocked()
	}
	switch e.Status {
	case StatusWorking:
		l.sealBuildLocked(e)
		l.upsertLive(eventKey(e), e)
		l.paintLocked()
		l.startSpinLocked()
	case StatusDone, StatusError:
		l.sealBuildLocked(e)
		l.removeLive(eventKey(e))
		l.commitLocked(l.format(e))
	default:
		l.commitLocked(l.format(e))
	}
}

// projectUpRow is the live "supabase up" line logAround opens for Compose Up.
// Attached up follows container logs until the context ends, so that row
// would spin for the whole session. Container events are the loading UI.
func projectUpRow(e LogEvent) bool {
	return e.Kind == KindCompose && e.Status == StatusWorking && e.Op == "up" && e.ID != "" &&
		e.Text == "" && e.Detail == "" && e.Total == 0
}

func (l *Logger) hidingProjectUp(e LogEvent) bool {
	return l.tty && !projectUpRow(e)
}

func (l *Logger) retireProjectUpLocked() {
	dropped := false
	for _, k := range append([]string(nil), l.live...) {
		if projectUpRow(l.rows[k]) {
			l.removeLive(k)
			dropped = true
		}
	}
	if dropped {
		l.paintLocked()
	}
}

func (l *Logger) repeatDone(e LogEvent) bool {
	if l.lastDone == nil {
		l.lastDone = map[string]string{}
	}
	key := e.ID + "\x00" + e.Parent
	sig := e.Op + "\x00" + e.Detail
	if l.lastDone[key] == sig {
		return true
	}
	l.lastDone[key] = sig
	return false
}

func eventKey(e LogEvent) string {
	id := e.ID
	parent := e.Parent
	if parent == id {
		parent = ""
	}
	if id != "" && parent != "" {
		return parent + "\x00" + id
	}
	if id != "" {
		return id
	}
	if parent != "" {
		return parent
	}
	return string(e.Kind) + "\x00" + e.Op
}

func (l *Logger) upsertLive(key string, e LogEvent) {
	if _, ok := l.rows[key]; !ok {
		l.live = append(l.live, key)
	}
	l.rows[key] = e
}

func (l *Logger) removeLive(key string) {
	if _, ok := l.rows[key]; !ok {
		return
	}
	delete(l.rows, key)
	out := l.live[:0]
	for _, k := range l.live {
		if k != key {
			out = append(out, k)
		}
	}
	l.live = out
}

func (l *Logger) format(e LogEvent) string {
	glyph := statusGlyph(e.Status)
	if l.tty && e.Status == StatusWorking && len(spinFrames) > 0 {
		glyph = spinFrames[l.spin%len(spinFrames)]
	}
	return formatEvent(e, l.colorOn(), l.timestamps, glyph)
}

// sealBuildLocked commits a build step that this event replaces.
// The loading header stays live until the first real step arrives.
func (l *Logger) sealBuildLocked(next LogEvent) {
	if next.Kind != KindBuild || next.Status == StatusError {
		return
	}
	key := eventKey(next)
	prev, ok := l.rows[key]
	if !ok {
		return
	}
	if prev.Op == next.Op && prev.Detail == next.Detail && prev.Text == next.Text {
		return
	}
	if prev.Op == "building" && prev.Detail == "" && prev.Text == "" {
		return
	}
	prev.Status = StatusDone
	l.removeLive(key)
	l.commitLocked(l.format(prev))
}

func (l *Logger) clearLiveLocked(key string) {
	if _, ok := l.rows[key]; !ok {
		return
	}
	l.removeLive(key)
	if l.tty {
		l.paintLocked()
	}
}

func (l *Logger) paintLocked() {
	if l.drawn > 0 {
		_, _ = io.WriteString(l.out, "\x1b["+itoa(l.drawn)+"A")
	}
	width := l.termWidth()
	pending := l.pending
	l.pending = nil
	written := 0
	for _, line := range pending {
		l.writeLineLocked(line, width)
		written += visualRows(line, width)
	}
	rows := 0
	for _, k := range l.live {
		line := l.format(l.rows[k])
		l.writeLineLocked(line, width)
		rows += visualRows(line, width)
	}
	extra := l.drawn - written - rows
	if extra < 0 {
		extra = 0
	}
	for i := 0; i < extra; i++ {
		_, _ = io.WriteString(l.out, "\r\x1b[2K\n")
	}
	if extra > 0 {
		_, _ = io.WriteString(l.out, "\x1b["+itoa(extra)+"A")
	}
	l.drawn = rows
}

func (l *Logger) commitLocked(line string) {
	l.pending = append(l.pending, line)
	l.paintLocked()
}

// writeLineLocked clears every visual row a wrapping line will occupy, then
// writes it. Clearing only the cursor row leaves the tail of the spinner
// glued to the end of a long log line.
func (l *Logger) writeLineLocked(line string, width int) {
	h := visualRows(line, width)
	if h <= 1 {
		_, _ = io.WriteString(l.out, "\r\x1b[2K"+line+"\n")
		return
	}
	var b strings.Builder
	for i := 0; i < h; i++ {
		b.WriteString("\r\x1b[2K")
		if i+1 < h {
			b.WriteByte('\n')
		}
	}
	b.WriteString("\x1b[" + itoa(h-1) + "A\r")
	b.WriteString(line)
	b.WriteByte('\n')
	_, _ = io.WriteString(l.out, b.String())
}

func (l *Logger) termWidth() int {
	if l.cols < 0 {
		return 0
	}
	if l.cols > 0 {
		return l.cols
	}
	f, ok := l.out.(*os.File)
	if !ok {
		return 0
	}
	w, _, err := term.GetSize(int(f.Fd()))
	if err != nil || w < 1 {
		return 0
	}
	return w
}

func (l *Logger) startSpinLocked() {
	if l.spinning || !l.tty {
		return
	}
	interval := l.spinEvery
	if interval == 0 {
		interval = 100 * time.Millisecond
	}
	if interval < 0 {
		return
	}
	l.spinning = true
	go l.spinLoop(interval)
}

func (l *Logger) spinLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		l.mu.Lock()
		if len(l.live) == 0 {
			l.spinning = false
			l.mu.Unlock()
			return
		}
		l.spin++
		l.paintLocked()
		l.mu.Unlock()
	}
}

func (l *Logger) advanceSpinner() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.tty || len(l.live) == 0 {
		return
	}
	l.spin++
	l.paintLocked()
}

func loggerFrom(ctx context.Context) *Logger {
	if ctx == nil {
		return nil
	}
	log, _ := ctx.Value(loggerContextKey{}).(*Logger)
	return log
}

// LoggerFrom returns the logger [Logger.Context] attached, or nil.
func LoggerFrom(ctx context.Context) *Logger {
	return loggerFrom(ctx)
}

// NewFrameLogger emits events to sink and does not write a human transcript.
// The session command uses it so log and status lines become NDJSON frames.
func NewFrameLogger(sink func(LogEvent)) *Logger {
	l := NewLogger(OutputStream(io.Discard), Color(false), Progress("plain"))
	l.frames = true
	l.sink = sink
	return l
}

func adoptLog(ctx context.Context, log **Logger) *Logger {
	if log != nil && *log != nil {
		return *log
	}
	found := loggerFrom(ctx)
	if log != nil && found != nil {
		*log = found
	}
	return found
}

func logAround(ctx context.Context, log *Logger, kind LogKind, op, id string, run func(context.Context) error) error {
	if log == nil {
		log = loggerFrom(ctx)
	}
	if log != nil {
		ctx = log.Context(ctx)
		log.Emit(LogEvent{Kind: kind, Op: op, ID: id, Status: StatusWorking})
	}
	err := run(ctx)
	if log == nil {
		return err
	}
	e := LogEvent{Kind: kind, Op: op, ID: id, Status: StatusDone}
	if err != nil {
		e.Status = StatusError
		e.Detail = err.Error()
	}
	log.Emit(e)
	return err
}

func emitPhase(log *Logger, kind LogKind, op, id string, status LogStatus, detail string) {
	if log == nil {
		return
	}
	log.Emit(LogEvent{Kind: kind, Op: op, ID: id, Detail: detail, Status: status})
}

func failPhase(log *Logger, kind LogKind, op, id string, err error) {
	if log == nil || err == nil {
		return
	}
	emitPhase(log, kind, op, id, StatusError, err.Error())
	log.settlePhase(id, StatusError)
}

func finishPhase(log *Logger, kind LogKind, op, id string) {
	if log == nil {
		return
	}
	if !log.tty {
		log.Emit(LogEvent{Kind: kind, Op: op, ID: id, Status: StatusDone})
		return
	}
	log.mu.Lock()
	open := phaseOpenLocked(log, id)
	log.mu.Unlock()
	if open {
		log.Emit(LogEvent{Kind: kind, Op: op, ID: id, Status: StatusDone})
	}
	log.settlePhase(id, StatusDone)
}

func phaseOpenLocked(l *Logger, id string) bool {
	if _, ok := l.rows[id]; ok {
		return true
	}
	prefix := id + "\x00"
	for k := range l.rows {
		if strings.HasPrefix(k, prefix) {
			return true
		}
	}
	return false
}

// settlePhase commits leftover live rows for id so a finished pull or build
// does not leave a spinner on screen.
func (l *Logger) settlePhase(id string, status LogStatus) {
	if l == nil || !l.tty || id == "" {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	prefix := id + "\x00"
	keys := make([]string, 0, len(l.live))
	for _, k := range l.live {
		if k == id || strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	for _, k := range keys {
		e := l.rows[k]
		if e.Status == StatusWorking {
			e.Status = status
		}
		l.removeLive(k)
		l.commitLocked(l.format(e))
	}
}

func isTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	st, err := f.Stat()
	if err != nil {
		return false
	}
	return st.Mode()&os.ModeCharDevice != 0
}

var (
	_ composeapi.EventProcessor = (*Logger)(nil)
	_ composeapi.LogConsumer    = (*Logger)(nil)
)
