package terminal

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"os"
	"sync"
	"testing"
	"time"
)

func TestNewSessionRejectsNil(t *testing.T) {
	t.Parallel()
	in, _ := pipeFiles(t)
	local, _ := netPipe(t)
	cases := []struct {
		name   string
		in     *os.File
		out    io.Writer
		conn   io.WriteCloser
		reader io.Reader
		want   string
	}{
		{name: "stdin", out: io.Discard, conn: local, reader: local, want: "terminal: nil stdin"},
		{name: "stdout", in: in, conn: local, reader: local, want: "terminal: nil stdout"},
		{name: "conn", in: in, out: io.Discard, reader: local, want: "terminal: nil connection"},
		{name: "reader", in: in, out: io.Discard, conn: local, want: "terminal: nil reader"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewSession(tc.in, tc.out, tc.conn, tc.reader)
			if err == nil || err.Error() != tc.want {
				t.Fatalf("err = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestSizeRequiresTerminal(t *testing.T) {
	t.Parallel()
	s := newPipeSession(t, io.Discard)
	_, err := s.Size()
	if err == nil || err.Error() != "terminal: stdin is not a terminal" {
		t.Fatalf("err = %v", err)
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	t.Parallel()
	s := newPipeSession(t, io.Discard)
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background(), nil); !errors.Is(err, ErrClosed) {
		t.Fatalf("Start after Close = %v, want ErrClosed", err)
	}
	if err := (*Session)(nil).Close(); err != nil {
		t.Fatal(err)
	}
}

func TestStartRejectsNilContext(t *testing.T) {
	t.Parallel()
	s := newPipeSession(t, io.Discard)
	if err := s.Start(nil, nil); err == nil || err.Error() != "terminal: nil context" {
		t.Fatalf("err = %v", err)
	}
}

func TestStartCopiesOutputUntilRemoteCloses(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	s, inW, peer := newPipeSessionEnds(t, &out)
	defer inW.Close()

	go func() {
		_, _ = peer.Write([]byte("hello"))
		_ = peer.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := s.Start(ctx, nil); err != nil {
		t.Fatal(err)
	}
	if out.String() != "hello" {
		t.Fatalf("out = %q, want hello", out.String())
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestStartHalfClosesOnStdinEOF(t *testing.T) {
	t.Parallel()
	inR, inW := pipeFiles(t)
	local, peer := netPipe(t)
	writeClosed := make(chan struct{})
	conn := &halfCloser{Conn: local, closed: writeClosed}
	s, err := NewSession(inR, io.Discard, conn, local)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- s.Start(ctx, nil) }()

	got := make(chan []byte, 1)
	go func() {
		buf := make([]byte, 4)
		n, err := io.ReadFull(peer, buf)
		if err != nil {
			got <- nil
			return
		}
		got <- append([]byte(nil), buf[:n]...)
	}()

	if _, err := inW.Write([]byte("ping")); err != nil {
		t.Fatal(err)
	}
	if err := inW.Close(); err != nil {
		t.Fatal(err)
	}

	select {
	case <-writeClosed:
	case <-ctx.Done():
		t.Fatal("stdin EOF did not half-close the connection")
	}
	select {
	case buf := <-got:
		if string(buf) != "ping" {
			t.Fatalf("forwarded = %q, want ping", buf)
		}
	case <-ctx.Done():
		t.Fatal("stdin bytes were not forwarded")
	}

	if err := peer.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("Start did not return after the remote closed")
	}
}

func TestStartStopsWhenContextCanceled(t *testing.T) {
	t.Parallel()
	reading := make(chan struct{})
	inR, _ := pipeFiles(t)
	local, _ := netPipe(t)
	s, err := NewSession(inR, io.Discard, local, &signalReader{Reader: local, reading: reading})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.Start(ctx, nil) }()
	select {
	case <-reading:
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not read output")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Start = %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return after cancel")
	}
}

func TestCloseUnblocksStart(t *testing.T) {
	t.Parallel()
	reading := make(chan struct{})
	inR, _ := pipeFiles(t)
	local, _ := netPipe(t)
	s, err := NewSession(inR, io.Discard, local, &signalReader{Reader: local, reading: reading})
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() { done <- s.Start(context.Background(), nil) }()
	select {
	case <-reading:
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not read output")
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if !errors.Is(err, ErrClosed) {
			t.Fatalf("Start = %v, want ErrClosed", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Close did not unblock Start")
	}
}

func TestStartOnce(t *testing.T) {
	t.Parallel()
	reading := make(chan struct{})
	inR, _ := pipeFiles(t)
	local, _ := netPipe(t)
	s, err := NewSession(inR, io.Discard, local, &signalReader{Reader: local, reading: reading})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- s.Start(ctx, nil) }()
	select {
	case <-reading:
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not read output")
	}
	if err := s.Start(context.Background(), nil); err == nil || err.Error() != "terminal: session already started" {
		t.Fatalf("second Start = %v", err)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return")
	}
}

func TestStartResizeRequiresTerminal(t *testing.T) {
	t.Parallel()
	s := newPipeSession(t, io.Discard)
	called := false
	err := s.Start(context.Background(), func(Size) error {
		called = true
		return nil
	})
	if err == nil || !bytes.Contains([]byte(err.Error()), []byte("stdin is not a terminal")) {
		t.Fatalf("err = %v", err)
	}
	if called {
		t.Fatal("resize called")
	}
}

func TestStartReturnsOutputError(t *testing.T) {
	t.Parallel()
	inR, _ := pipeFiles(t)
	local, _ := netPipe(t)
	s, err := NewSession(inR, io.Discard, local, errReader{err: errors.New("boom")})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = s.Start(ctx, nil)
	if err == nil || !bytes.Contains([]byte(err.Error()), []byte("boom")) {
		t.Fatalf("err = %v", err)
	}
}

func TestWatchResizeReturnsWhenContextCanceled(t *testing.T) {
	t.Parallel()
	s := newPipeSession(t, io.Discard)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan error, 1)
	go func() {
		done <- s.watchResize(ctx, func(Size) error {
			t.Error("resize called")
			return nil
		})
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("watchResize did not return")
	}
}

func TestStartCanceledContextDoesNotConsumeSession(t *testing.T) {
	t.Parallel()
	s, inW, peer := newPipeSessionEnds(t, io.Discard)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.Start(ctx, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("Start = %v, want context.Canceled", err)
	}

	if err := inW.Close(); err != nil {
		t.Fatal(err)
	}
	if err := peer.Close(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := s.Start(ctx, nil); err != nil {
		t.Fatal(err)
	}
}

func newPipeSession(t *testing.T, out io.Writer) *Session {
	t.Helper()
	s, inW, peer := newPipeSessionEnds(t, out)
	t.Cleanup(func() {
		inW.Close()
		peer.Close()
	})
	return s
}

func newPipeSessionEnds(t *testing.T, out io.Writer) (*Session, *os.File, net.Conn) {
	t.Helper()
	inR, inW := pipeFiles(t)
	local, peer := netPipe(t)
	s, err := NewSession(inR, out, local, local)
	if err != nil {
		t.Fatal(err)
	}
	return s, inW, peer
}

func pipeFiles(t *testing.T) (*os.File, *os.File) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		r.Close()
		w.Close()
	})
	return r, w
}

func netPipe(t *testing.T) (net.Conn, net.Conn) {
	t.Helper()
	a, b := net.Pipe()
	t.Cleanup(func() {
		a.Close()
		b.Close()
	})
	return a, b
}

type halfCloser struct {
	net.Conn
	closed chan struct{}
	once   sync.Once
}

func (h *halfCloser) CloseWrite() error {
	h.once.Do(func() { close(h.closed) })
	return nil
}

type signalReader struct {
	io.Reader
	reading chan struct{}
	once    sync.Once
}

func (s *signalReader) Read(p []byte) (int, error) {
	s.once.Do(func() { close(s.reading) })
	return s.Reader.Read(p)
}

type errReader struct{ err error }

func (e errReader) Read([]byte) (int, error) { return 0, e.err }
