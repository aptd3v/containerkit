package taskflow

import "time"

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
	// Task is the step. Step names are not unique across subflows; compare Task.
	Task *Task
	Name string
	Kind EventKind
	// Err is set for EventError and EventRetry.
	Err error
	// Attempt is the 1-based attempt that just failed. Set for EventRetry.
	Attempt  int
	Duration time.Duration
}

func (f *Flow) emit(e Event) {
	if f == nil {
		return
	}
	f.mu.Lock()
	watch := f.watch
	f.mu.Unlock()
	if watch != nil {
		watch(e)
	}
}

func (f *Flow) emitLog(step string, p []byte) {
	if f == nil || len(p) == 0 {
		return
	}
	f.mu.Lock()
	logfn := f.logfn
	f.mu.Unlock()
	if logfn != nil {
		logfn(step, p)
	}
}

type logWriter struct {
	f    *Flow
	step string
}

func (w logWriter) Write(p []byte) (int, error) {
	if w.f == nil {
		return len(p), nil
	}
	cp := append([]byte(nil), p...)
	w.f.emitLog(w.step, cp)
	return len(p), nil
}
