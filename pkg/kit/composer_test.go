package kit_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/aptd3v/containerkit/pkg/kit"
)

func TestNewComposerNilOption(t *testing.T) {
	t.Parallel()
	_, err := kit.NewComposer(kit.Host("http://localhost"), nil)
	if err == nil {
		t.Fatal("expected nil option")
	}
}

func TestNewComposerConstructsService(t *testing.T) {
	t.Parallel()
	transport := roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
	})
	log := kit.NewLogger(kit.OutputStream(io.Discard), kit.Color(false), kit.Progress("quiet"))
	c, err := kit.NewComposer(
		kit.Host("http://localhost"),
		kit.HTTPClient(&http.Client{Transport: transport}),
		kit.Log(log),
		kit.MaxConcurrency(2),
		kit.OutputStream(io.Discard),
	)
	if err != nil {
		t.Fatal(err)
	}
	if c == nil {
		t.Fatal("nil composer")
	}
	var _ kit.Composer = c
}

func TestNewComposerSettersCompile(t *testing.T) {
	t.Parallel()
	var (
		_ kit.ComposerClientOption = kit.FromEnv()
		_ kit.ComposerClientOption = kit.Host("http://localhost")
		_ kit.ComposerClientOption = kit.Log(kit.NewLogger())
		_ kit.ComposerClientOption = kit.MaxConcurrency(4)
		_ kit.ComposerClientOption = kit.DryRun(true)
		_ kit.ComposerClientOption = kit.OutputStream(io.Discard)
	)
}
