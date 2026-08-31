package terminal

import (
	"io"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewSessionRequiresTTY(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer r.Close()
	defer w.Close()
	_, err = NewSession(r, nil, nil)
	require.Error(t, err)
}

func TestSessionCloseAndStart(t *testing.T) {
	s := &Session{stdin: os.Stdin}
	require.NoError(t, s.Close())

	pr, pw, err := os.Pipe()
	require.NoError(t, err)
	defer pr.Close()
	defer pw.Close()

	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()

	sess := &Session{
		stdin:    pr,
		hijacked: a,
		reader:   io.NopCloser(b),
	}
	done := make(chan error, 1)
	go func() { done <- sess.Start() }()
	require.NoError(t, pw.Close())
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return after stdin EOF")
	}

	s2 := &Session{stdin: os.Stdin, resizeCh: make(chan MonitorSize, 1)}
	_, _, _ = s2.GetSize()
}
