package kit

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	composeerrdefs "github.com/compose-spec/compose-go/v2/errdefs"
	cerrdefs "github.com/containerd/errdefs"
	composeapi "github.com/docker/compose/v5/pkg/api"
)

func TestErrorClassifiesEngineAndComposeSentinels(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		cause error
		class error
		is    func(error) bool
	}{
		{"engine not found", cerrdefs.ErrNotFound, ErrNotFound, IsNotFound},
		{"compose not found", composeapi.ErrNotFound, ErrNotFound, IsNotFound},
		{"compose-go not found", composeerrdefs.ErrNotFound, ErrNotFound, IsNotFound},
		{"conflict", cerrdefs.ErrConflict, ErrConflict, IsConflict},
		{"unauthorized", cerrdefs.ErrUnauthenticated, ErrUnauthorized, IsUnauthorized},
		{"forbidden", composeapi.ErrForbidden, ErrPermissionDenied, IsPermissionDenied},
		{"already exists", composeapi.ErrAlreadyExists, ErrAlreadyExists, IsAlreadyExists},
		{"unavailable", cerrdefs.ErrUnavailable, ErrUnavailable, IsUnavailable},
		{"not implemented", composeapi.ErrNotImplemented, ErrNotImplemented, IsNotImplemented},
		{"canceled", composeapi.ErrCanceled, ErrCanceled, IsCanceled},
		{"login failed", ErrLoginFailed, ErrLoginFailed, IsLoginFailed},
		{"invalid project", composeerrdefs.ErrInvalid, ErrInvalidArgument, IsInvalidArgument},
		{"unsupported attribute", composeerrdefs.ErrUnsupported, ErrUnsupportedAttribute, IsUnsupportedAttribute},
		{"incompatible", composeerrdefs.ErrIncompatible, ErrIncompatible, IsIncompatible},
		{"required config", errRequired, ErrInvalidArgument, IsInvalidArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := cfgErr("engine", "alpine:latest", "PullImage", tt.cause)
			if !errors.Is(err, tt.class) {
				t.Fatalf("errors.Is class: %v", err)
			}
			if !tt.is(err) {
				t.Fatalf("helper: %v", err)
			}
			if ClassOf(err) != tt.class && !errors.Is(ClassOf(err), tt.class) {
				t.Fatalf("ClassOf = %v want %v", ClassOf(err), tt.class)
			}
			if err.Hint() == "" && tt.class == ErrNotFound {
				t.Fatal("expected hint")
			}
			if err.StatusCode() == 0 && tt.class == ErrNotFound {
				t.Fatal("expected http status")
			}
		})
	}
}

func TestErrorAsTypedSDKPayloads(t *testing.T) {
	t.Parallel()
	tmpl := InvalidTemplateError{Template: "${"}
	err := cfgErr("compose", "import", "Import", tmpl)
	var got InvalidTemplateError
	if !errors.As(err, &got) || got.Template != "${" {
		t.Fatalf("As InvalidTemplateError: %v", err)
	}
	if !IsInvalidArgument(err) {
		t.Fatalf("template should classify as invalid argument: %v", err)
	}

	miss := MissingRequiredError{Variable: "TAG", Reason: "unset"}
	err = cfgErr("compose", "import", "Import", miss)
	var gotMiss MissingRequiredError
	if !errors.As(err, &gotMiss) || gotMiss.Variable != "TAG" {
		t.Fatalf("As MissingRequiredError: %v", err)
	}
	if err.Path != "TAG" {
		t.Fatalf("path = %q", err.Path)
	}

	js := &JSONError{Code: 404, Message: "no such image"}
	err = cfgErr("engine", "alpine:latest", "PullImage", js)
	var gotJS *JSONError
	if !errors.As(err, &gotJS) || gotJS.Code != 404 {
		t.Fatalf("As JSONError: %v", err)
	}
	if !IsNotFound(err) {
		t.Fatalf("json 404 should be not found: %v", err)
	}
}

func TestErrorFromWaitAndResponse(t *testing.T) {
	t.Parallel()
	if err := ErrorFromResponse(ErrorResponse{}); err != nil {
		t.Fatalf("empty response: %v", err)
	}
	err := ErrorFromResponse(ErrorResponse{Message: "boom"})
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("response: %v", err)
	}
	if err := ErrorFromWait(WaitResponse{}); err != nil {
		t.Fatalf("empty wait: %v", err)
	}
	err = ErrorFromWait(WaitResponse{
		StatusCode: 137,
		Error:      &WaitExitError{Message: "killed"},
	})
	var exit *ExitError
	if !errors.As(err, &exit) || exit.StatusCode != 137 || exit.Message != "killed" {
		t.Fatalf("ExitError: %+v (%v)", exit, err)
	}
}

func TestErrorDebugFormat(t *testing.T) {
	t.Parallel()
	err := cfgErr("engine", "alpine:latest", "PullImage", cerrdefs.ErrNotFound)
	dump := fmt.Sprintf("%+v", err)
	for _, want := range []string{"kit.Error", "receiver: engine", "op: PullImage", "resource: alpine:latest", "class: not found", "hint:"} {
		if !strings.Contains(dump, want) {
			t.Fatalf("missing %q in %+v", want, dump)
		}
	}
}

func TestJSONStreamErrorIsReturned(t *testing.T) {
	t.Parallel()
	log := NewLogger(OutputStream(io.Discard), Color(false), Progress("plain"))
	body := `{"errorDetail":{"code":404,"message":"no such image"},"error":"no such image"}` + "\n"
	err := log.drainJSON(context.Background(), strings.NewReader(body), KindPull, "alpine:latest")
	if err == nil {
		t.Fatal("expected json stream error")
	}
	if !IsNotFound(err) {
		t.Fatalf("classified: %v", err)
	}
	var js *JSONError
	if !errors.As(err, &js) || js.Message != "no such image" {
		t.Fatalf("As JSONError: %v", err)
	}
}

func TestWaitSurfacesSetupErrors(t *testing.T) {
	t.Parallel()
	c := NewContainer(Image("alpine:latest"))
	res := c.Wait(context.Background(), nil, "")
	select {
	case err := <-res.Error:
		if err == nil {
			t.Fatal("expected wait error")
		}
		if !IsInvalidArgument(err) {
			t.Fatalf("wait setup: %v", err)
		}
	default:
		t.Fatal("expected error on wait channel")
	}
}

func TestWrappedErrorKeepsPath(t *testing.T) {
	t.Parallel()
	inner := cfgErr("", "ports", "", errInvalid)
	wrapped := fmt.Errorf("load: %w", inner)
	got := wrapPath("service", wrapped)
	var ce *Error
	if !errors.As(got, &ce) {
		t.Fatalf("As: %v", got)
	}
	if ce.Path != "service.ports" {
		t.Fatalf("path = %q", ce.Path)
	}
	if !strings.Contains(got.Error(), "load:") {
		t.Fatalf("wrap dropped: %v", got)
	}

	filled := cfgErr("compose", "ports", "Import", fmt.Errorf("load: %w", &Error{err: errInvalid}))
	if filled.Receiver != "compose" || filled.Path != "ports" || filled.Setting != "Import" {
		t.Fatalf("cfgErr = receiver %q path %q setting %q", filled.Receiver, filled.Path, filled.Setting)
	}
	if !errors.Is(filled, errInvalid) {
		t.Fatalf("cause: %v", filled)
	}
}

func TestCanceledClassifies(t *testing.T) {
	t.Parallel()
	err := cfgErr("engine", "pull", "PullImage", context.Canceled)
	if !IsCanceled(err) {
		t.Fatalf("canceled: %v", err)
	}
}
