package kit

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	errNilOption      = errors.New("nil option")
	errRequired       = errors.New("required")
	errInvalid        = errors.New("invalid")
	errUnsupported    = errors.New("unsupported mapping")
	errEmpty          = errors.New("empty")
	errSecretRedacted = errors.New("invalid secret value")
)

// Error identifies the receiving object, operation, and resource.
// SDK failures are classified onto [ErrNotFound] and the other sentinels so
// [errors.Is] / [errors.As] work without importing Engine or Compose.
// Secret values are redacted. Unwrap preserves the original cause.
type Error struct {
	Receiver string
	Path     string
	Setting  string
	secret   bool
	class    error
	status   int
	hint     string
	err      error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	msg := "invalid configuration"
	if e.err != nil {
		if e.secret {
			msg = errSecretRedacted.Error()
		} else {
			msg = e.err.Error()
		}
	}
	recv := e.Receiver
	if recv == "" {
		recv = "config"
	}
	path := strings.TrimPrefix(e.Path, ".")
	if path == "" {
		path = e.Setting
	}
	loc := recv
	if path != "" {
		loc = recv + " " + path
	}
	parts := []string{loc}
	if e.Setting != "" && path != e.Setting && !strings.HasSuffix(path, e.Setting) {
		parts = append(parts, e.Setting)
	}
	if cls := classLabel(e.class); cls != "" && !strings.Contains(strings.ToLower(msg), cls) {
		parts = append(parts, cls)
	}
	parts = append(parts, msg)
	return strings.Join(parts, ": ")
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

// Is reports whether target matches this error's class or cause.
func (e *Error) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}
	if e.class != nil && (e.class == target || errors.Is(e.class, target)) {
		return true
	}
	return false
}

// Class is the SDK error class ([ErrNotFound], [ErrConflict], …).
func (e *Error) Class() error {
	if e == nil {
		return nil
	}
	return e.class
}

// Hint is a short debugging note for %+v / [Error.Hint].
func (e *Error) Hint() string {
	if e == nil {
		return ""
	}
	return e.hint
}

// StatusCode is the HTTP status implied by the class, when known.
func (e *Error) StatusCode() int {
	if e == nil {
		return 0
	}
	return e.status
}

// Format renders a one-line error, or a multi-line dump with %+v.
func (e *Error) Format(s fmt.State, verb rune) {
	if e == nil {
		_, _ = io.WriteString(s, "<nil>")
		return
	}
	if verb == 'v' && s.Flag('+') {
		_, _ = io.WriteString(s, e.debug())
		return
	}
	if verb == 'q' {
		fmt.Fprintf(s, "%q", e.Error())
		return
	}
	_, _ = io.WriteString(s, e.Error())
}

func (e *Error) debug() string {
	var b strings.Builder
	fmt.Fprintf(&b, "kit.Error: %s\n", e.Error())
	if e.Receiver != "" {
		fmt.Fprintf(&b, "  receiver: %s\n", e.Receiver)
	}
	if e.Setting != "" {
		fmt.Fprintf(&b, "  op: %s\n", e.Setting)
	}
	if path := strings.TrimPrefix(e.Path, "."); path != "" {
		fmt.Fprintf(&b, "  resource: %s\n", path)
	}
	if e.class != nil {
		fmt.Fprintf(&b, "  class: %s\n", e.class.Error())
	}
	if e.status != 0 {
		fmt.Fprintf(&b, "  status: %d\n", e.status)
	}
	if e.hint != "" {
		fmt.Fprintf(&b, "  hint: %s\n", e.hint)
	}
	if e.err != nil && !e.secret {
		fmt.Fprintf(&b, "  cause: %v", e.err)
	} else if e.secret {
		fmt.Fprintf(&b, "  cause: %s", errSecretRedacted)
	}
	return strings.TrimRight(b.String(), "\n")
}

func cfgErr(receiver, path, setting string, err error) *Error {
	var ce *Error
	if errors.As(err, &ce) && ce != nil {
		if ce.Receiver == "" {
			ce.Receiver = receiver
		}
		if ce.Path == "" || ce.Path == "." {
			ce.Path = path
		}
		if ce.Setting == "" {
			ce.Setting = setting
		}
		ce.classify()
		return ce
	}
	e := &Error{Receiver: receiver, Path: path, Setting: setting, err: err}
	e.classify()
	return e
}

func secretErr(receiver, path, setting string, err error) *Error {
	e := &Error{Receiver: receiver, Path: path, Setting: setting, secret: true, err: err}
	e.classify()
	return e
}

func wrapReceiver(receiver string, err error) error {
	if err == nil {
		return nil
	}
	var ce *Error
	if errors.As(err, &ce) && ce != nil {
		if ce.Receiver == "" {
			ce.Receiver = receiver
		}
		ce.classify()
		return err
	}
	e := &Error{Receiver: receiver, Path: ".", err: err}
	e.classify()
	return e
}

func join(errs ...error) error {
	return errors.Join(errs...)
}

func addErr(errs []error, err error) []error {
	if err == nil {
		return errs
	}
	return append(errs, err)
}

func wrapPath(path string, err error) error {
	if err == nil {
		return nil
	}
	var ce *Error
	if errors.As(err, &ce) && ce != nil {
		if ce.Path == "" || ce.Path == "." {
			ce.Path = path
		} else if !strings.HasPrefix(ce.Path, path) {
			ce.Path = path + "." + strings.TrimPrefix(ce.Path, ".")
		}
		ce.classify()
		return err
	}
	e := &Error{Path: path, err: err}
	e.classify()
	return e
}

func classLabel(class error) string {
	if class == nil {
		return ""
	}
	return class.Error()
}
