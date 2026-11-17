package kit

import (
	"context"
	"sync"

	composeapi "github.com/aptd3v/containerkit/internal/compose/api/types"
)

// logBridge is the Compose event processor installed on every client.
// Compose captures the processor when the service is constructed.
// Start receives the call context, so a logger passed to Up is visible here
// even when Connect was given none. On and Done do not receive a context;
// they use the logger Start pushed.
//
// One Compose operation at a time per client. Overlapping calls share this stack.
type logBridge struct {
	fallback *Logger
	mu       sync.Mutex
	stack    []logFrame
}

type logFrame struct {
	log     *Logger
	forward bool
}

func newLogBridge(fallback *Logger) *logBridge {
	return &logBridge{fallback: fallback}
}

// Start records the call logger. Start and Done are forwarded only when the
// context has no logger. logAround already brackets that call, including the
// error text. A Connect logger, with no context logger, still receives them.
func (b *logBridge) Start(ctx context.Context, operation string) {
	if b == nil {
		return
	}
	log := LoggerFrom(ctx)
	forward := false
	if log == nil {
		log = b.fallback
		forward = log != nil
	}
	b.mu.Lock()
	b.stack = append(b.stack, logFrame{log: log, forward: forward})
	b.mu.Unlock()
	if forward {
		log.Start(ctx, operation)
	}
}

// On emits resource rows to the logger for the operation in progress.
func (b *logBridge) On(events ...composeapi.Resource) {
	if b == nil {
		return
	}
	if log := b.current(); log != nil {
		log.On(events...)
	}
}

// Done pops the operation started by Start.
func (b *logBridge) Done(operation string, failed bool) {
	if b == nil {
		return
	}
	b.mu.Lock()
	var frame logFrame
	if n := len(b.stack); n > 0 {
		frame = b.stack[n-1]
		b.stack = b.stack[:n-1]
	} else {
		frame = logFrame{log: b.fallback, forward: b.fallback != nil}
	}
	b.mu.Unlock()
	if frame.forward && frame.log != nil {
		frame.log.Done(operation, failed)
	}
}

func (b *logBridge) current() *Logger {
	b.mu.Lock()
	defer b.mu.Unlock()
	if n := len(b.stack); n > 0 && b.stack[n-1].log != nil {
		return b.stack[n-1].log
	}
	return b.fallback
}

var _ composeapi.EventProcessor = (*logBridge)(nil)
