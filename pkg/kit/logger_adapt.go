package kit

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	composeapi "github.com/aptd3v/containerkit/internal/compose/api/types"
	controlapi "github.com/moby/buildkit/api/services/control"
	"github.com/moby/moby/api/types/jsonstream"
	"google.golang.org/protobuf/proto"
)

type jsonMsg struct {
	Stream       string               `json:"stream,omitempty"`
	Status       string               `json:"status,omitempty"`
	Progress     *jsonstream.Progress `json:"progressDetail,omitempty"`
	ID           string               `json:"id,omitempty"`
	Error        *jsonstream.Error    `json:"errorDetail,omitempty"`
	ErrorMessage string               `json:"error,omitempty"`
	Aux          json.RawMessage      `json:"aux,omitempty"`
}

func (l *Logger) drainJSON(ctx context.Context, r io.Reader, kind LogKind, parent string) error {
	if r == nil {
		return nil
	}
	if c, ok := r.(io.Closer); ok {
		defer c.Close()
	}
	dec := json.NewDecoder(r)
	for {
		if err := ctx.Err(); err != nil {
			return cfgErr("engine", parent, string(kind), err)
		}
		var msg jsonMsg
		err := dec.Decode(&msg)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return cfgErr("engine", parent, string(kind), fmt.Errorf("progress stream: %w", err))
		}
		if msg.Error != nil && msg.Error.Message != "" {
			if l != nil {
				if e, ok := eventFromJSON(kind, parent, msg); ok {
					l.Emit(e)
				}
			}
			return cfgErr("engine", parent, string(kind), msg.Error)
		}
		if l != nil {
			if msg.ID == "moby.buildkit.trace" {
				l.emitBuildkitTrace(kind, parent, msg.Aux)
				continue
			}
			if e, ok := eventFromJSON(kind, parent, msg); ok {
				l.Emit(e)
			}
		}
	}
}

// emitBuildkitTrace turns a BuildKit status payload into log and status events.
// The engine sends that payload as a base64 protobuf aux message, which has no
// human-readable status or stream fields.
func (l *Logger) emitBuildkitTrace(kind LogKind, parent string, aux json.RawMessage) {
	if l == nil || len(aux) == 0 {
		return
	}
	var encoded string
	if err := json.Unmarshal(aux, &encoded); err != nil || encoded == "" {
		return
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return
	}
	var status controlapi.StatusResponse
	if err := proto.Unmarshal(raw, &status); err != nil {
		return
	}
	if l.vertexNames == nil {
		l.vertexNames = map[string]string{}
	}
	if l.vertexDone == nil {
		l.vertexDone = map[string]bool{}
	}
	for _, vertex := range status.GetVertexes() {
		name := strings.TrimSpace(vertex.GetName())
		if name != "" && l.vertexNames[vertex.GetDigest()] == "" {
			l.vertexNames[vertex.GetDigest()] = name
			l.Emit(LogEvent{Kind: kind, Op: "step", ID: parent, Text: name, Status: StatusWorking, Parent: parent})
		}
		if vertex.GetCompleted() != nil && vertex.GetDigest() != "" && !l.vertexDone[vertex.GetDigest()] {
			l.vertexDone[vertex.GetDigest()] = true
			errText := strings.TrimSpace(vertex.GetError())
			if errText == "" && !vertex.GetCached() {
				continue
			}
			ev := LogEvent{Kind: kind, Op: "step", ID: firstNonEmpty(name, l.vertexNames[vertex.GetDigest()]), Status: StatusDone, Parent: parent}
			if vertex.GetCached() {
				ev.Detail = "CACHED"
			}
			if errText != "" {
				ev.Status = StatusError
				ev.Detail = errText
			}
			if ev.ID != "" {
				l.Emit(ev)
			}
		}
	}
	for _, entry := range status.GetLogs() {
		name := l.vertexNames[entry.GetVertex()]
		for _, line := range strings.Split(strings.TrimRight(string(entry.GetMsg()), "\r\n"), "\n") {
			line = strings.TrimRight(line, "\r")
			if strings.TrimSpace(line) == "" {
				continue
			}
			l.Emit(LogEvent{Kind: KindLog, Op: "stdout", ID: name, Detail: line, Status: StatusInfo, Parent: parent})
		}
	}
	for _, entry := range status.GetStatuses() {
		label := strings.TrimSpace(entry.GetName())
		if label == "" {
			continue
		}
		who := firstNonEmpty(entry.GetID(), l.vertexNames[entry.GetVertex()], parent)
		ev := LogEvent{Kind: kind, Op: "progress", ID: who, Text: label, Status: StatusWorking, Parent: parent}
		if entry.GetTotal() > 0 {
			ev.Current = entry.GetCurrent()
			ev.Total = entry.GetTotal()
		}
		l.Emit(ev)
	}
}

func eventFromJSON(kind LogKind, parent string, msg jsonMsg) (LogEvent, bool) {
	if msg.Error != nil && msg.Error.Message != "" {
		return LogEvent{
			Kind:   kind,
			Op:     "error",
			ID:     firstNonEmpty(msg.ID, parent),
			Parent: parent,
			Detail: msg.Error.Message,
			Status: StatusError,
		}, true
	}
	if msg.ErrorMessage != "" {
		return LogEvent{
			Kind:   kind,
			Op:     "error",
			ID:     firstNonEmpty(msg.ID, parent),
			Parent: parent,
			Detail: msg.ErrorMessage,
			Status: StatusError,
		}, true
	}
	if len(msg.Aux) > 0 {
		var aux struct {
			ID string `json:"ID"`
		}
		if json.Unmarshal(msg.Aux, &aux) == nil && aux.ID != "" {
			return LogEvent{
				Kind:   kind,
				Op:     "built",
				ID:     firstNonEmpty(parent, aux.ID),
				Parent: parent,
				Detail: aux.ID,
				Status: StatusDone,
			}, true
		}
	}
	stream := strings.TrimSpace(msg.Stream)
	status := strings.TrimSpace(msg.Status)
	if stream == "" && status == "" && msg.ID == "" {
		return LogEvent{}, false
	}
	e := LogEvent{
		Kind:   kind,
		Op:     jsonOp(status, stream),
		ID:     firstNonEmpty(msg.ID, parent),
		Parent: parent,
		Text:   firstNonEmpty(status, stream),
		Status: jsonStatus(status, stream),
	}
	if stream != "" && status == "" {
		e.Detail = stream
	}
	if msg.Progress != nil && msg.Progress.Total > 0 {
		e.Current = msg.Progress.Current
		e.Total = msg.Progress.Total
		e.Status = StatusWorking
	}
	return e, e.Op != "" || e.Detail != "" || e.ID != ""
}

func jsonOp(status, stream string) string {
	s := strings.ToLower(strings.TrimSpace(status))
	switch {
	case strings.HasPrefix(s, "pulling fs layer"), strings.HasPrefix(s, "pulling from"), strings.HasPrefix(s, "pulling"):
		return "pulling"
	case strings.HasPrefix(s, "waiting"):
		return "waiting"
	case strings.HasPrefix(s, "downloading"):
		return "downloading"
	case strings.HasPrefix(s, "download complete"):
		return "downloaded"
	case strings.HasPrefix(s, "extracting"):
		return "extracting"
	case strings.HasPrefix(s, "pull complete"), strings.HasPrefix(s, "already exists"):
		return "pulled"
	case strings.HasPrefix(s, "pushing"):
		return "pushing"
	case strings.HasPrefix(s, "pushed"), strings.HasPrefix(s, "layer already exists"):
		return "pushed"
	case strings.HasPrefix(s, "preparing"):
		return "preparing"
	case strings.HasPrefix(s, "digest:"):
		return "digest"
	case strings.HasPrefix(s, "status:"):
		if strings.Contains(s, "up to date") {
			return "exists"
		}
		return "pulled"
	}
	line := strings.TrimSpace(stream)
	switch {
	case strings.HasPrefix(line, "Step "):
		return "step"
	case strings.HasPrefix(line, "--->"), strings.HasPrefix(line, " --->"):
		return "layer"
	case line != "":
		return "stream"
	}
	if s != "" {
		return s
	}
	return "progress"
}

func jsonStatus(status, stream string) LogStatus {
	s := strings.ToLower(strings.TrimSpace(status))
	switch {
	case strings.Contains(s, "complete"), strings.Contains(s, "already exists"),
		strings.Contains(s, "downloaded newer"), strings.Contains(s, "up to date"),
		strings.Contains(s, "pushed"), strings.HasPrefix(s, "digest:"):
		return StatusDone
	case strings.HasPrefix(s, "error"):
		return StatusError
	}
	line := strings.TrimSpace(stream)
	if strings.HasPrefix(line, "Step ") {
		return StatusWorking
	}
	if line != "" {
		return StatusInfo
	}
	return StatusWorking
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// Start implements compose EventProcessor.
func (l *Logger) Start(_ context.Context, operation string) {
	if l == nil {
		return
	}
	l.Emit(LogEvent{
		Kind:   KindCompose,
		Op:     strings.ToLower(strings.TrimSpace(operation)),
		Status: StatusWorking,
	})
}

// On implements compose EventProcessor.
func (l *Logger) On(events ...composeapi.Resource) {
	if l == nil {
		return
	}
	for _, e := range events {
		st := StatusWorking
		var status composeapi.EventStatus = e.Status
		switch status {
		case composeapi.Done:
			st = StatusDone
		case composeapi.Warning:
			st = StatusWarn
		case composeapi.Error:
			st = StatusError
		}
		op, text, detail := composeEventText(st, e)
		l.Emit(LogEvent{
			Kind:    KindCompose,
			Op:      op,
			ID:      e.ID,
			Parent:  e.ParentID,
			Text:    text,
			Detail:  detail,
			Current: e.Current,
			Total:   e.Total,
			Status:  st,
		})
	}
}

func composeEventText(st LogStatus, e composeapi.Resource) (op, text, detail string) {
	op = strings.ToLower(strings.TrimSpace(e.Text))
	if op == "" {
		op = strings.ToLower(strings.TrimSpace(e.StatusText()))
	}
	text = e.Text
	detail = e.Details
	if st != StatusError {
		return op, text, detail
	}
	msg := strings.TrimSpace(e.Text)
	if strings.TrimSpace(detail) != "" {
		msg = strings.TrimSpace(detail)
	}
	if !strings.Contains(msg, " ") {
		return op, text, detail
	}
	if strings.TrimSpace(detail) == "" {
		detail = strings.TrimSpace(e.Text)
	}
	return "error", "Error", detail
}

// Done implements compose EventProcessor.
// Compose calls this with true when the operation returned an error
// (the flag is err != nil, not success).
func (l *Logger) Done(operation string, failed bool) {
	if l == nil {
		return
	}
	st := StatusDone
	if failed {
		st = StatusError
	}
	l.Emit(LogEvent{
		Kind:   KindCompose,
		Op:     strings.ToLower(strings.TrimSpace(operation)),
		Status: st,
	})
}

// Log implements compose LogConsumer.
func (l *Logger) Log(containerName, message string) {
	l.Emit(LogEvent{
		Kind:   KindLog,
		Op:     "stdout",
		ID:     containerName,
		Detail: strings.TrimRight(message, "\r\n"),
		Status: StatusInfo,
	})
}

// Err implements compose LogConsumer.
func (l *Logger) Err(containerName, message string) {
	l.Emit(LogEvent{
		Kind:   KindLog,
		Op:     "stderr",
		ID:     containerName,
		Detail: strings.TrimRight(message, "\r\n"),
		Status: StatusWarn,
	})
}

// Status implements compose LogConsumer.
func (l *Logger) Status(container, msg string) {
	l.Emit(LogEvent{
		Kind:   KindCompose,
		Op:     "status",
		ID:     container,
		Detail: strings.TrimRight(msg, "\r\n"),
		Status: StatusInfo,
	})
}
