package kit

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestCopyMultiplexedSplitsStdoutAndStderr(t *testing.T) {
	t.Parallel()
	var src bytes.Buffer
	writeFrame(&src, 1, "out")
	writeFrame(&src, 2, "err")
	var stdout, stderr bytes.Buffer
	n, err := CopyMultiplexed(&stdout, &stderr, &src)
	if err != nil {
		t.Fatal(err)
	}
	if n != 6 || stdout.String() != "out" || stderr.String() != "err" {
		t.Fatalf("n=%d stdout=%q stderr=%q", n, stdout.String(), stderr.String())
	}
}

func writeFrame(dst *bytes.Buffer, stream byte, payload string) {
	var header [8]byte
	header[0] = stream
	binary.BigEndian.PutUint32(header[4:], uint32(len(payload)))
	dst.Write(header[:])
	dst.WriteString(payload)
}
