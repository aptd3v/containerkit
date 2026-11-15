package kit

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
)

// Session keeps one Docker client for a stream of plan calls.
type Session struct {
	mu  sync.Mutex
	cli *Client
	key string
	log *Logger
}

// Exec validates plan JSON, dials on the first call, and reuses that client
// while connect options stay the same.
func (s *Session) Exec(ctx context.Context, data []byte) ([]byte, error) {
	if s == nil {
		return nil, cfgErr("tsx", "session", "Exec", fmt.Errorf("%w: session", errRequired))
	}
	p := ImportPlan(data)
	if err := p.Validate(); err != nil {
		return nil, err
	}
	cli, err := s.client(p.raw.Connect)
	if err != nil {
		return nil, err
	}
	if s.log != nil {
		ctx = s.log.Context(ctx)
	}
	return p.runActions(ctx, cli)
}

func (s *Session) client(connect map[string]any) (*Client, error) {
	key, err := connectKey(connect)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cli != nil && s.key == key {
		return s.cli, nil
	}
	opts, err := decodeConnectOptions(connect)
	if err != nil {
		return nil, err
	}
	cli, err := Connect(opts...)
	if err != nil {
		return nil, err
	}
	var closeErr error
	if s.cli != nil {
		closeErr = s.cli.Close()
	}
	s.cli = cli
	s.key = key
	if closeErr != nil {
		return nil, closeErr
	}
	return cli, nil
}

// Close closes the session client.
func (s *Session) Close() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cli == nil {
		return nil
	}
	err := s.cli.Close()
	s.cli = nil
	s.key = ""
	return err
}

func connectKey(connect map[string]any) (string, error) {
	if len(connect) == 0 {
		return "", nil
	}
	b, err := json.Marshal(connect)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

type sessionFrame struct {
	ID   json.RawMessage `json:"id,omitempty"`
	Type string          `json:"type"`
	Plan json.RawMessage `json:"plan,omitempty"`
}

type sessionResult struct {
	ID    json.RawMessage `json:"id,omitempty"`
	Type  string          `json:"type"`
	OK    bool            `json:"ok"`
	Value json.RawMessage `json:"value,omitempty"`
}

type sessionError struct {
	ID       json.RawMessage `json:"id,omitempty"`
	Type     string          `json:"type"`
	Class    string          `json:"class"`
	Status   int             `json:"status,omitempty"`
	Message  string          `json:"message"`
	Hint     string          `json:"hint,omitempty"`
	Receiver string          `json:"receiver,omitempty"`
	Path     string          `json:"path,omitempty"`
	ExitCode *int64          `json:"exitCode,omitempty"`
}

type frameWriter struct {
	mu  sync.Mutex
	enc *json.Encoder
	err error
}

func (w *frameWriter) write(v any) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.err != nil {
		return w.err
	}
	if err := w.enc.Encode(v); err != nil {
		w.err = err
		return err
	}
	return nil
}

// ErrorClass is the stable class string for a kit error.
// Callers branch on this value. The message stays redacted.
func ErrorClass(err error) string {
	if err == nil {
		return ""
	}
	switch {
	case IsLoginRequired(err):
		return "loginRequired"
	case IsLoginFailed(err):
		return "loginFailed"
	case IsNotFound(err):
		return "notFound"
	case IsAlreadyExists(err):
		return "alreadyExists"
	case IsConflict(err):
		return "conflict"
	case IsUnauthorized(err):
		return "unauthorized"
	case IsPermissionDenied(err):
		return "permissionDenied"
	case IsInvalidProject(err):
		return "invalidProject"
	case IsWrongContextType(err):
		return "wrongContextType"
	case IsParsingFailed(err):
		return "parsingFailed"
	case IsUnsupportedFlag(err):
		return "unsupportedFlag"
	case IsUnsupportedAttribute(err), errors.Is(err, errUnsupported):
		return "unsupported"
	case IsIncompatible(err):
		return "incompatible"
	case IsDisabled(err):
		return "disabled"
	case IsInvalidArgument(err):
		return "invalidArgument"
	case IsUnavailable(err), IsConnectionFailed(err):
		return "unavailable"
	case IsCanceled(err):
		return "canceled"
	case IsNotImplemented(err):
		return "notImplemented"
	case IsResourceExhausted(err):
		return "resourceExhausted"
	case IsFailedPrecondition(err):
		return "failedPrecondition"
	case IsAborted(err):
		return "aborted"
	case IsOutOfRange(err):
		return "outOfRange"
	case IsDataLoss(err):
		return "dataLoss"
	case IsNotModified(err):
		return "notModified"
	case IsInternal(err):
		return "internal"
	case IsUnknown(err):
		return "unknown"
	default:
		return "unknown"
	}
}

func errorReport(err error) sessionError {
	rep := sessionError{
		Type:    "error",
		Class:   ErrorClass(err),
		Message: err.Error(),
	}
	var ke *Error
	if errors.As(err, &ke) && ke != nil {
		rep.Status = ke.StatusCode()
		rep.Hint = ke.Hint()
		rep.Receiver = ke.Receiver
		path := strings.TrimPrefix(ke.Path, ".")
		if path == "" || path == "." {
			path = ke.Setting
		}
		rep.Path = path
	}
	var exit *ExitError
	if errors.As(err, &exit) && exit != nil {
		code := exit.StatusCode
		rep.ExitCode = &code
	}
	if rep.Class == "" {
		rep.Class = "unknown"
	}
	return rep
}

func eventFrame(e LogEvent) map[string]any {
	kind := "status"
	if e.Kind == KindLog {
		kind = "log"
	}
	source := e.ID
	if source == "" {
		source = e.Op
	}
	line := e.Detail
	if line == "" {
		line = e.Text
	}
	return map[string]any{
		"type":   "event",
		"kind":   kind,
		"source": source,
		"stream": e.Op,
		"line":   line,
	}
}

// ServeSession reads NDJSON calls from in and writes result, error, and event
// frames to out. Diagnostics go to errw. The process stays up until in
// reaches EOF or ctx is canceled.
func ServeSession(ctx context.Context, in io.Reader, out, errw io.Writer) int {
	w := &frameWriter{enc: json.NewEncoder(out)}
	var log *Logger
	log = NewFrameLogger(func(e LogEvent) {
		if err := w.write(eventFrame(e)); err != nil {
			return
		}
	})
	s := &Session{log: log}
	defer s.Close()
	return serve(ctx, in, errw, w, func(ctx context.Context, data []byte) ([]byte, error) {
		return s.Exec(ctx, data)
	})
}

func serve(ctx context.Context, in io.Reader, errw io.Writer, w *frameWriter, exec func(context.Context, []byte) ([]byte, error)) int {
	if ctx == nil {
		ctx = context.Background()
	}
	lines := make(chan string)
	readErr := make(chan error, 1)
	go func() {
		defer close(lines)
		sc := bufio.NewScanner(in)
		sc.Buffer(make([]byte, 64*1024), 16*1024*1024)
		for sc.Scan() {
			select {
			case <-ctx.Done():
				return
			case lines <- sc.Text():
			}
		}
		if err := sc.Err(); err != nil {
			readErr <- err
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return 0
		case err := <-readErr:
			fmt.Fprintln(errw, err)
			return 1
		case line, ok := <-lines:
			if !ok {
				return 0
			}
			if strings.TrimSpace(line) == "" {
				continue
			}
			if err := handleFrame(ctx, w, line, exec); err != nil {
				fmt.Fprintln(errw, err)
				return 1
			}
		}
	}
}

func handleFrame(ctx context.Context, w *frameWriter, line string, exec func(context.Context, []byte) ([]byte, error)) error {
	var frame sessionFrame
	if err := json.Unmarshal([]byte(line), &frame); err != nil {
		return w.write(sessionError{
			Type:    "error",
			Class:   "invalidArgument",
			Message: "invalid session frame",
		})
	}
	if frame.Type != "exec" {
		rep := sessionError{
			ID:      frame.ID,
			Type:    "error",
			Class:   "invalidArgument",
			Message: "unknown session frame",
		}
		return w.write(rep)
	}
	value, err := exec(ctx, frame.Plan)
	if err != nil {
		rep := errorReport(err)
		rep.ID = frame.ID
		return w.write(rep)
	}
	raw := json.RawMessage(`{"ok":true}`)
	if json.Valid(value) {
		raw = json.RawMessage(value)
	}
	return w.write(sessionResult{
		ID:    frame.ID,
		Type:  "result",
		OK:    true,
		Value: raw,
	})
}
