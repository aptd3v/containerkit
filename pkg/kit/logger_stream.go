package kit

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	composeapi "github.com/aptd3v/containerkit/internal/compose/api/types"
)

const maxLogFrame = 1 << 20

// loggingLogs is a container log reader that emits LogEvents as it is read.
// Read returns the original bytes, including multiplexed headers.
type loggingLogs struct {
	src       io.ReadCloser
	log       *Logger
	id        string
	raw       []byte
	out       []byte
	stdout    []byte
	stderr    []byte
	errStatus LogStatus
	mode      int // 0 unknown, 1 multiplex, 2 raw stdout
	closed    bool
}

func newLoggingLogs(src io.ReadCloser, log *Logger, id string) *loggingLogs {
	return &loggingLogs{src: src, log: log, id: id}
}

func (r *loggingLogs) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	for len(r.out) == 0 {
		err := r.pull()
		if len(r.out) > 0 {
			break
		}
		if err != nil {
			return 0, err
		}
	}
	n := copy(p, r.out)
	r.out = r.out[n:]
	return n, nil
}

func (r *loggingLogs) Close() error {
	if r == nil || r.closed {
		return nil
	}
	r.closed = true
	r.scan(true)
	if r.src == nil {
		return nil
	}
	err := r.src.Close()
	r.src = nil
	return err
}

func (r *loggingLogs) pull() error {
	if r.src == nil {
		r.scan(true)
		return io.EOF
	}
	buf := make([]byte, 32*1024)
	n, err := r.src.Read(buf)
	if n > 0 {
		r.raw = append(r.raw, buf[:n]...)
	}
	r.scan(errors.Is(err, io.EOF))
	if n > 0 && err == nil {
		return nil
	}
	return err
}

func (r *loggingLogs) scan(eof bool) {
	if r.mode == 0 {
		if len(r.raw) < 8 && !eof {
			return
		}
		if len(r.raw) >= 8 {
			if _, ok := multiplexHeader(r.raw[:8]); ok {
				r.mode = 1
			} else {
				r.mode = 2
			}
		} else {
			r.mode = 2
		}
	}
	if r.mode == 2 {
		if len(r.raw) > 0 {
			r.out = append(r.out, r.raw...)
			r.emit(r.raw, "stdout", &r.stdout, StatusInfo)
			r.raw = nil
		}
		if eof {
			r.flush()
		}
		return
	}
	for len(r.raw) >= 8 {
		size, ok := multiplexHeader(r.raw[:8])
		if !ok {
			r.out = append(r.out, r.raw...)
			r.emit(r.raw, "stdout", &r.stdout, StatusInfo)
			r.raw = nil
			break
		}
		total := 8 + size
		if len(r.raw) < total {
			break
		}
		frame := r.raw[:total]
		payload := frame[8:]
		op, status := streamEvent(frame[0])
		rest := &r.stdout
		if op == "stderr" {
			rest = &r.stderr
			if status == StatusError {
				r.errStatus = StatusError
			}
		}
		r.out = append(r.out, frame...)
		r.emit(payload, op, rest, status)
		r.raw = r.raw[total:]
	}
	if eof {
		if len(r.raw) > 0 {
			r.out = append(r.out, r.raw...)
			r.raw = nil
		}
		r.flush()
	}
}

func (r *loggingLogs) emit(payload []byte, op string, rest *[]byte, status LogStatus) {
	if r.log == nil || len(payload) == 0 {
		return
	}
	*rest = append(*rest, payload...)
	for {
		i := bytes.IndexByte(*rest, '\n')
		if i < 0 {
			return
		}
		line := string((*rest)[:i])
		*rest = (*rest)[i+1:]
		r.log.Emit(LogEvent{
			Kind:   KindLog,
			Op:     op,
			ID:     r.id,
			Detail: trimCR(line),
			Status: status,
		})
	}
}

func (r *loggingLogs) flush() {
	if r.log == nil {
		return
	}
	if len(r.stdout) > 0 {
		r.log.Emit(LogEvent{Kind: KindLog, Op: "stdout", ID: r.id, Detail: trimCR(string(r.stdout)), Status: StatusInfo})
		r.stdout = nil
	}
	if len(r.stderr) > 0 {
		status := r.errStatus
		if status == "" {
			status = StatusWarn
		}
		r.log.Emit(LogEvent{Kind: KindLog, Op: "stderr", ID: r.id, Detail: trimCR(string(r.stderr)), Status: status})
		r.stderr = nil
	}
}

func trimCR(s string) string {
	if len(s) > 0 && s[len(s)-1] == '\r' {
		return s[:len(s)-1]
	}
	return s
}

func streamEvent(stream byte) (string, LogStatus) {
	switch stream {
	case 2:
		return "stderr", StatusWarn
	case 3:
		return "stderr", StatusError
	default:
		return "stdout", StatusInfo
	}
}

func multiplexHeader(b []byte) (int, bool) {
	if len(b) < 8 || b[0] > 3 || b[1] != 0 || b[2] != 0 || b[3] != 0 {
		return 0, false
	}
	n := int(binary.BigEndian.Uint32(b[4:8]))
	if n < 0 || n > maxLogFrame {
		return 0, false
	}
	return n, true
}

// consumeContainerLogs reads a container log stream to EOF.
// A reader from [Container.Logs] emits while it is read. The plan runner
// uses this so the session logger sees those lines and the result is {ok:true}.
func consumeContainerLogs(r io.ReadCloser) error {
	if r == nil {
		return nil
	}
	_, err := io.Copy(io.Discard, r)
	cerr := r.Close()
	if err != nil {
		return err
	}
	return cerr
}

type teeLogConsumer struct {
	first  composeapi.LogConsumer
	second composeapi.LogConsumer
}

func (t teeLogConsumer) Log(name, message string) {
	if t.first != nil {
		t.first.Log(name, message)
	}
	if t.second != nil {
		t.second.Log(name, message)
	}
}

func (t teeLogConsumer) Err(name, message string) {
	if t.first != nil {
		t.first.Err(name, message)
	}
	if t.second != nil {
		t.second.Err(name, message)
	}
}

func (t teeLogConsumer) Status(name, message string) {
	if t.first != nil {
		t.first.Status(name, message)
	}
	if t.second != nil {
		t.second.Status(name, message)
	}
}

var _ composeapi.LogConsumer = teeLogConsumer{}
