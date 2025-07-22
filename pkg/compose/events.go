package compose

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
)

type Events struct {
	// Core event fields
	Time    string `json:"time"`
	Type    string `json:"type"`
	Action  string `json:"action"`
	ID      string `json:"id"`
	Service string `json:"service"`
	// Attributes map containing event-specific information
	Attributes map[string]string `json:"attributes"`

	// Additional fields that may be present in some events
	Scope  string `json:"scope,omitempty"`
	Status string `json:"status,omitempty"`
	From   string `json:"from,omitempty"`
}

type eventsWriter struct {
	ctx    context.Context
	ch     chan Events
	errCh  chan error
	buffer bytes.Buffer
}

func (w *eventsWriter) Write(p []byte) (n int, err error) {
	n, err = w.buffer.Write(p)
	if err != nil {
		return n, err
	}

	// Scan line by line
	for {
		line, err := w.buffer.ReadBytes('\n')
		if err == io.EOF {
			// Incomplete line; keep in buffer
			break
		} else if err != nil {
			return n, err
		}

		var e Events
		if err := json.Unmarshal(line, &e); err == nil {
			select {
			case <-w.ctx.Done():
				w.buffer.Reset()
				return n, w.ctx.Err()
			case w.ch <- e:
			}
		} else {
			select {
			case <-w.ctx.Done():
				w.buffer.Reset()
				return n, w.ctx.Err()
			case w.errCh <- NewComposeEventsError(err):

			}
		}
	}

	return n, nil
}

func newEventsWriter(ctx context.Context, ch chan Events, errCh chan error) io.Writer {
	return &eventsWriter{ctx: ctx, ch: ch, errCh: errCh}
}
