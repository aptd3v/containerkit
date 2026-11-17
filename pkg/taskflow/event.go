package taskflow

import (
	"time"

	"github.com/aptd3v/containerkit/pkg/kit"
)

// Status is the terminal outcome of a step. The zero value means the step has
// not finished, including while it is running or retrying.
type Status string

const (
	// StatusDone means the step finished without error.
	StatusDone Status = "done"
	// StatusSkip means the step did not run because a predecessor failed.
	StatusSkip Status = "skip"
	// StatusError means the step failed, panicked, or its context was already canceled.
	StatusError Status = "error"
)

// ReasonAlreadyPresent is Event.Reason when SkipIfPresent finds every tag.
const ReasonAlreadyPresent = "already_present"

// EventKind is a step lifecycle signal sent to Run.Watch.
type EventKind string

const (
	EventStart EventKind = "start"
	EventDone  EventKind = "done"
	EventSkip  EventKind = "skip"
	EventError EventKind = "error"
	// EventRetry is sent when an attempt failed and another follows.
	EventRetry EventKind = "retry"
)

// Event is one step lifecycle update. Watch may be called from several goroutines.
type Event struct {
	// Task is the step. Step names are not unique across subflows.
	Task *Task
	Name string
	Kind EventKind
	// Err is set for EventError and EventRetry.
	Err error
	// Attempt is the 1-based attempt that just failed. Set for EventRetry.
	Attempt int
	// Reason explains a done step that did not do the usual work.
	// ReasonAlreadyPresent means SkipIfPresent found every tag.
	Reason   string
	Duration time.Duration
}

func (f *Flow) emit(e Event) {
	if f == nil {
		return
	}
	f.mu.Lock()
	watch := f.watch
	ctx := f.ctx
	name := f.name
	f.mu.Unlock()
	if watch != nil {
		watch(e)
	}
	log := kit.LoggerFrom(ctx)
	if log == nil {
		return
	}
	status := kit.StatusInfo
	detail := e.Name
	switch e.Kind {
	case EventStart:
		status = kit.StatusWorking
	case EventDone:
		status = kit.StatusDone
	case EventSkip:
		status = kit.StatusWarn
		detail = "skip"
	case EventRetry, EventError:
		status = kit.StatusError
		if e.Kind == EventRetry {
			status = kit.StatusWarn
		}
		if e.Err != nil {
			detail = e.Err.Error()
		}
	}
	log.Emit(kit.LogEvent{
		Kind:   kit.KindEngine,
		Op:     e.Name,
		ID:     name,
		Detail: detail,
		Status: status,
	})
}

func (f *Flow) emitLog(step string, p []byte) {
	if f == nil || len(p) == 0 {
		return
	}
	f.mu.Lock()
	logfn := f.logfn
	ctx := f.ctx
	name := f.name
	f.mu.Unlock()
	if logfn != nil {
		logfn(step, p)
	}
	if log := kit.LoggerFrom(ctx); log != nil {
		log.Emit(kit.LogEvent{
			Kind:   kit.KindLog,
			Op:     step,
			ID:     name,
			Detail: string(p),
			Status: kit.StatusInfo,
		})
	}
}

type logWriter struct {
	f    *Flow
	step string
}

func (w logWriter) Write(p []byte) (int, error) {
	if w.f == nil || len(p) == 0 {
		return len(p), nil
	}
	cp := append([]byte(nil), p...)
	w.f.emitLog(w.step, cp)
	return len(p), nil
}
