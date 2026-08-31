package client

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func muxStream(stream byte, payload string) []byte {
	hdr := make([]byte, 8)
	hdr[0] = stream
	binary.BigEndian.PutUint32(hdr[4:], uint32(len(payload)))
	return append(hdr, payload...)
}

func TestLogCopier(t *testing.T) {
	var stdout, stderr bytes.Buffer
	lc := NewLogCopier(&stdout, nil)
	require.Equal(t, lc.stdout, lc.stderr)

	lc = NewLogCopier(&stdout, &stderr)
	data := append(muxStream(1, "out\n"), muxStream(2, "err\n")...)
	n, err := lc.Copy(bytes.NewReader(data))
	require.NoError(t, err)
	require.Greater(t, n, int64(0))
	require.Equal(t, "out\n", stdout.String())
	require.Equal(t, "err\n", stderr.String())

	stdout.Reset()
	stderr.Reset()
	_, err = lc.CopyWithPrefix(bytes.NewReader(muxStream(1, "line\n")), "SOUT:", "SERR:")
	require.NoError(t, err)
	require.Contains(t, stdout.String(), "SOUT:")
}

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errors.New("write fail") }

func TestPrefixWriterError(t *testing.T) {
	w := &prefixWriter{writer: failWriter{}, prefix: "p:"}
	_, err := w.Write([]byte("x"))
	require.Error(t, err)
	_, _ = io.Copy(io.Discard, strings.NewReader(""))
}
