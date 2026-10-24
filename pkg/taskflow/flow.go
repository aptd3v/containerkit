package taskflow

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/aptd3v/containerkit/pkg/client"
	"github.com/aptd3v/containerkit/pkg/client/response"
	"github.com/aptd3v/containerkit/pkg/kit"
	gtf "github.com/noneback/go-taskflow"
)

// Priority is scheduler queue order. Zero leaves go-taskflow's default.
type Priority uint

const (
	PriorityHigh Priority = iota + 1
	PriorityNormal
	PriorityLow
)

type kind int

const (
	kindTask kind = iota + 1
	kindContainer
	kindBuild
	kindCompose
	kindSubflow
	kindCondition
	kindIf
)

type gtfAdder interface {
	NewTask(name string, f func()) *gtf.Task
	NewSubflow(name string, instantiate func(*gtf.Subflow)) *gtf.Task
	NewCondition(name string, predict func() uint) *gtf.Task
}

// scope is where a step is added: the root graph or one subflow.
type scope struct {
	flow   *Flow
	adder  gtfAdder
	parent *Task
}

// Flow is a named task graph. One Run per Flow.
type Flow struct {
	name string
	cli  *client.Client
	tf   *gtf.TaskFlow

	mu      sync.Mutex
	tasks   []*Task
	created []string
	compose []*kit.Compose
	ctx     context.Context
	writer  io.Writer
	ran     bool
	watch   func(Event)
	logfn   func(string, []byte)
}

// Subflow is a nested graph. It has the same step constructors as Flow.
type Subflow struct {
	flow *Flow
	gtf  *gtf.Subflow
	task *Task
}

// Task is one node in a Flow or Subflow.
type Task struct {
	flow  *Flow
	adder gtfAdder
	name  string
	kind  kind
	gtf   *gtf.Task

	preds []*Task
	succs []*Task

	// thenNode and elseNode are the two fixed successors of an If step.
	// fallback is the hidden last successor of a Condition; an out-of-range
	// index routes there instead of panicking inside go-taskflow.
	thenNode *gtf.Task
	elseNode *gtf.Task
	fallback *gtf.Task

	always  bool
	retries int

	container    *kit.Container
	containerRun *ContainerRun
	build        kit.Build
	buildOpt     *client.ImageBuild
	composeApp   *kit.Compose
	composeFn    func(context.Context, *kit.Compose) error
	taskFn       func(context.Context) error
	condFn       func(context.Context) uint
	ifFn         func(context.Context) bool

	mu       sync.Mutex
	children []*Task
	id       string
	exitCode int
	err      error
	skipped  bool
	tainted  bool
}

// Run is options for Flow.Run. Nil means defaults.
type Run struct {
	// Concurrency is the worker count. Zero means GOMAXPROCS; the value is
	// raised above the number of known subflows.
	Concurrency uint
	// Writer receives build output, container output, and Compose output.
	// Nil means os.Stdout.
	Writer io.Writer
	// Profile receives flamegraph text when the graph finishes. Nil disables the profiler.
	Profile io.Writer
	// Trace receives Chrome Trace JSON when the graph finishes. Nil disables the tracer.
	Trace io.Writer
	// Watch is called from worker goroutines as each step starts, finishes,
	// retries, skips, or fails.
	Watch func(Event)
	// Log is called with step-scoped output (image build streams, container
	// logs, compose output). May be called from worker goroutines with
	// partial writes.
	Log func(step string, p []byte)
}

// ContainerRun is options for a Container step. Nil means defaults.
type ContainerRun struct {
	// Logs receives the container's stdout and stderr after it exits. Nil
	// means Run.Writer. Use io.Discard to silence one step.
	Logs io.Writer
}

// New returns a Flow. name is the graph name and the flow label value.
// cli may be nil when the flow has only Task and Compose steps.
func New(name string, cli *client.Client) *Flow {
	if name == "" {
		name = "taskflow"
	}
	return &Flow{
		name: name,
		cli:  cli,
		tf:   gtf.NewTaskFlow(name),
		ctx:  context.Background(),
	}
}

// Name is the graph name passed to New.
func (f *Flow) Name() string {
	if f == nil {
		return ""
	}
	return f.name
}

func (f *Flow) scope() scope {
	return scope{flow: f, adder: f.tf}
}

func (s *Subflow) scope() (scope, bool) {
	if s == nil || s.flow == nil || s.gtf == nil {
		return scope{}, false
	}
	return scope{flow: s.flow, adder: s.gtf, parent: s.task}, true
}

// Task adds a function step.
func (f *Flow) Task(name string, fn func(context.Context) error) *Task {
	if f == nil {
		return nil
	}
	return f.scope().task(name, fn)
}

// Task adds a function step to the subflow.
func (s *Subflow) Task(name string, fn func(context.Context) error) *Task {
	sc, ok := s.scope()
	if !ok {
		return nil
	}
	return sc.task(name, fn)
}

// Subflow adds a nested graph. instantiate runs once, when the subflow starts.
// Successors of the subflow are skipped when any step inside it failed.
func (f *Flow) Subflow(name string, instantiate func(*Subflow)) *Task {
	if f == nil {
		return nil
	}
	return f.scope().subflow(name, instantiate)
}

// Subflow adds a nested graph inside this subflow.
func (s *Subflow) Subflow(name string, instantiate func(*Subflow)) *Task {
	sc, ok := s.scope()
	if !ok {
		return nil
	}
	return sc.subflow(name, instantiate)
}

// Condition adds an indexed branch. The returned index selects the successor
// in Precede order. A Condition runs even when a predecessor failed, which is
// what makes retry cycles work: a cycle needs a non-condition predecessor as
// the first entry, then a Condition on the back edge. An index with no
// successor fails the step instead of panicking; nothing runs after it.
// Use If to gate on a predecessor's success.
func (f *Flow) Condition(name string, predict func(context.Context) uint) *Task {
	if f == nil {
		return nil
	}
	return f.scope().condition(name, predict)
}

// Condition adds an indexed branch inside the subflow. See Flow.Condition.
func (s *Subflow) Condition(name string, predict func(context.Context) uint) *Task {
	sc, ok := s.scope()
	if !ok {
		return nil
	}
	return sc.condition(name, predict)
}

// If adds a boolean branch. Then successors run when pred returns true, Else
// successors when it returns false. A missing branch is a no-op. Unlike
// Condition, an If is skipped when a predecessor failed, and then neither
// branch runs.
func (f *Flow) If(name string, pred func(context.Context) bool) *Task {
	if f == nil {
		return nil
	}
	return f.scope().ifStep(name, pred)
}

// If adds a boolean branch inside the subflow. See Flow.If.
func (s *Subflow) If(name string, pred func(context.Context) bool) *Task {
	sc, ok := s.scope()
	if !ok {
		return nil
	}
	return sc.ifStep(name, pred)
}

func (sc scope) newTask(name string, k kind) *Task {
	t := &Task{flow: sc.flow, adder: sc.adder, name: name, kind: k}
	sc.flow.mu.Lock()
	sc.flow.tasks = append(sc.flow.tasks, t)
	sc.flow.mu.Unlock()
	if p := sc.parent; p != nil {
		p.mu.Lock()
		p.children = append(p.children, t)
		p.mu.Unlock()
	}
	return t
}

func (sc scope) task(name string, fn func(context.Context) error) *Task {
	t := sc.newTask(name, kindTask)
	t.taskFn = fn
	t.gtf = sc.adder.NewTask(name, func() {
		t.observe(func() error {
			if t.taskFn == nil {
				return nil
			}
			return t.taskFn(t.flow.context())
		})
	})
	return t
}

func (sc scope) subflow(name string, instantiate func(*Subflow)) *Task {
	t := sc.newTask(name, kindSubflow)
	f := sc.flow
	t.gtf = sc.adder.NewSubflow(name, func(sf *gtf.Subflow) {
		t.observe(func() error {
			if instantiate != nil {
				instantiate(&Subflow{flow: f, gtf: sf, task: t})
			}
			for _, c := range t.childList() {
				c.addFallback()
			}
			f.stamp()
			return nil
		})
	})
	return t
}

func (sc scope) condition(name string, predict func(context.Context) uint) *Task {
	t := sc.newTask(name, kindCondition)
	t.condFn = predict
	t.gtf = sc.adder.NewCondition(name, func() uint {
		choice := t.fallbackIndex()
		t.observe(func() error {
			if t.condFn == nil {
				return nil
			}
			got := t.condFn(t.flow.context())
			n := uint(len(t.succs))
			if got >= n {
				return fmt.Errorf("condition returned %d, has %d successors", got, n)
			}
			choice = got
			return nil
		})
		return choice
	})
	return t
}

func (sc scope) ifStep(name string, pred func(context.Context) bool) *Task {
	t := sc.newTask(name, kindIf)
	t.ifFn = pred
	t.thenNode = sc.adder.NewTask(name+":then", func() {})
	t.elseNode = sc.adder.NewTask(name+":else", func() {})
	t.gtf = sc.adder.NewCondition(name, func() uint {
		choice := uint(1)
		t.observe(func() error {
			if t.ifFn != nil && t.ifFn(t.flow.context()) {
				choice = 0
			}
			return nil
		})
		return choice
	})
	t.gtf.Precede(t.thenNode, t.elseNode)
	return t
}

// addFallback appends the hidden last successor of a Condition. It runs once
// all user edges exist: at Run for root steps, after instantiate for subflow steps.
func (t *Task) addFallback() {
	if t == nil || t.kind != kindCondition || t.fallback != nil || t.gtf == nil || t.adder == nil {
		return
	}
	t.fallback = t.adder.NewTask(t.name+":none", func() {})
	t.gtf.Precede(t.fallback)
}

func (t *Task) fallbackIndex() uint {
	if t.fallback == nil {
		return 0
	}
	return uint(len(t.succs))
}

func (f *Flow) context() context.Context {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.ctx == nil {
		return context.Background()
	}
	return f.ctx
}

func (f *Flow) hasLog() bool {
	if f == nil {
		return false
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.logfn != nil
}

func (f *Flow) out() io.Writer {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.writer == nil {
		return os.Stdout
	}
	return f.writer
}

func (f *Flow) trackID(id string) {
	if id == "" {
		return
	}
	f.mu.Lock()
	f.created = append(f.created, id)
	f.mu.Unlock()
}

func (f *Flow) trackCompose(app *kit.Compose) {
	if f == nil || app == nil {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, existing := range f.compose {
		if existing == app {
			return
		}
	}
	f.compose = append(f.compose, app)
}

func (f *Flow) composeList() []*kit.Compose {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]*kit.Compose(nil), f.compose...)
}

func (f *Flow) taskList() []*Task {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]*Task(nil), f.tasks...)
}

func (f *Flow) subflowCount() uint {
	var n uint
	for _, t := range f.taskList() {
		if t.kind == kindSubflow {
			n++
		}
	}
	return n
}

// Precede makes tasks depend on t. For a Condition, the order is the branch
// index. For an If, Precede is the same as Then.
func (t *Task) Precede(tasks ...*Task) *Task {
	if t == nil {
		return nil
	}
	if t.kind == kindIf {
		return t.link(t.thenNode, tasks)
	}
	return t.link(t.gtf, tasks)
}

// DependsOn makes t depend on tasks. It is the inverse of Precede.
func (t *Task) DependsOn(tasks ...*Task) *Task {
	if t == nil {
		return nil
	}
	for _, o := range tasks {
		if o != nil {
			o.Precede(t)
		}
	}
	return t
}

// Then adds successors that run when an If step returns true. On any other
// step it is the same as Precede.
func (t *Task) Then(tasks ...*Task) *Task {
	if t == nil {
		return nil
	}
	if t.kind != kindIf {
		return t.Precede(tasks...)
	}
	return t.link(t.thenNode, tasks)
}

// Else adds successors that run when an If step returns false. On any other
// step it does nothing.
func (t *Task) Else(tasks ...*Task) *Task {
	if t == nil || t.kind != kindIf {
		return t
	}
	return t.link(t.elseNode, tasks)
}

func (t *Task) link(from *gtf.Task, tasks []*Task) *Task {
	gtfTasks := make([]*gtf.Task, 0, len(tasks))
	for _, o := range tasks {
		if o == nil {
			continue
		}
		t.succs = append(t.succs, o)
		o.preds = append(o.preds, t)
		if o.gtf != nil {
			gtfTasks = append(gtfTasks, o.gtf)
		}
	}
	if from != nil && len(gtfTasks) > 0 {
		from.Precede(gtfTasks...)
	}
	return t
}

// Priority sets scheduler queue order. Zero is ignored.
func (t *Task) Priority(p Priority) *Task {
	if t == nil || t.gtf == nil {
		return t
	}
	switch p {
	case PriorityHigh:
		t.gtf.Priority(gtf.HIGH)
	case PriorityNormal:
		t.gtf.Priority(gtf.NORMAL)
	case PriorityLow:
		t.gtf.Priority(gtf.LOW)
	}
	return t
}

// Always runs t even when a predecessor failed or was skipped. Use it for
// teardown. The upstream failure still propagates: successors of t are
// skipped as if t had failed.
func (t *Task) Always() *Task {
	if t == nil {
		return nil
	}
	t.always = true
	return t
}

// Retry reruns a failing step up to n more times before it counts as failed.
// Each rerun emits EventRetry. Panics and a canceled context are not retried.
func (t *Task) Retry(n int) *Task {
	if t == nil {
		return nil
	}
	if n < 0 {
		n = 0
	}
	t.retries = n
	return t
}

// Name is the step name.
func (t *Task) Name() string {
	if t == nil {
		return ""
	}
	return t.name
}

// ID is the Engine container ID after a Container step runs.
func (t *Task) ID() string {
	if t == nil {
		return ""
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.id
}

// ExitCode is the container wait status after a Container step runs.
func (t *Task) ExitCode() int {
	if t == nil {
		return 0
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.exitCode
}

// Err is the step error after Run, if any. For a Subflow it reports only the
// subflow's own error; inner step errors are on the inner tasks and in the
// error returned by Run.
func (t *Task) Err() error {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.err
}

func (t *Task) begin() {
	t.mu.Lock()
	t.err = nil
	t.skipped = false
	t.tainted = false
	t.exitCode = 0
	t.id = ""
	t.mu.Unlock()
}

// resetResult clears per-attempt container state between retries.
func (t *Task) resetResult() {
	t.mu.Lock()
	t.exitCode = 0
	t.id = ""
	t.mu.Unlock()
}

func (t *Task) markSkipped() {
	t.mu.Lock()
	t.skipped = true
	t.mu.Unlock()
}

func (t *Task) markTainted() {
	t.mu.Lock()
	t.tainted = true
	t.mu.Unlock()
}

func (t *Task) setErr(err error) {
	if err == nil {
		return
	}
	t.mu.Lock()
	t.err = taskErr(t.name, err)
	t.mu.Unlock()
}

func (t *Task) setID(id string) {
	t.mu.Lock()
	t.id = id
	t.mu.Unlock()
}

func (t *Task) setExitCode(code int) {
	t.mu.Lock()
	t.exitCode = code
	t.mu.Unlock()
}

func (t *Task) childList() []*Task {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]*Task(nil), t.children...)
}

// failed reports whether t or any step inside it (for a Subflow) errored.
func (t *Task) failed() bool {
	t.mu.Lock()
	err := t.err
	t.mu.Unlock()
	if err != nil {
		return true
	}
	for _, c := range t.childList() {
		if c.failed() {
			return true
		}
	}
	return false
}

// blocks reports whether successors of t must be skipped.
func (t *Task) blocks() bool {
	t.mu.Lock()
	bad := t.skipped || t.tainted
	t.mu.Unlock()
	return bad || t.failed()
}

func (t *Task) predsFailed() bool {
	for _, p := range t.preds {
		if p.blocks() {
			return true
		}
	}
	return false
}

func (t *Task) event(k EventKind) Event {
	return Event{Task: t, Name: t.name, Kind: k}
}

func (t *Task) observe(fn func() error) {
	started := time.Now()
	finished := false
	defer func() {
		if r := recover(); r != nil {
			t.setErr(fmt.Errorf("panic: %v", r))
			if !finished {
				e := t.event(EventError)
				e.Err = t.Err()
				e.Duration = time.Since(started)
				t.flow.emit(e)
			}
		}
	}()
	t.begin()
	if t.predsFailed() {
		if !t.always && t.kind != kindCondition {
			t.markSkipped()
			t.flow.emit(t.event(EventSkip))
			finished = true
			return
		}
		if t.always {
			t.markTainted()
		}
	}
	ctx := t.flow.context()
	if err := ctx.Err(); err != nil {
		t.setErr(err)
		e := t.event(EventError)
		e.Err = t.Err()
		t.flow.emit(e)
		finished = true
		return
	}
	started = time.Now()
	t.flow.emit(t.event(EventStart))
	err := fn()
	for attempt := 1; err != nil && attempt <= t.retries && ctx.Err() == nil; attempt++ {
		e := t.event(EventRetry)
		e.Err = taskErr(t.name, err)
		e.Attempt = attempt
		e.Duration = time.Since(started)
		t.flow.emit(e)
		t.resetResult()
		started = time.Now()
		err = fn()
	}
	if err != nil {
		t.setErr(err)
		e := t.event(EventError)
		e.Err = t.Err()
		e.Duration = time.Since(started)
		t.flow.emit(e)
		finished = true
		return
	}
	e := t.event(EventDone)
	e.Duration = time.Since(started)
	t.flow.emit(e)
	finished = true
}

// Run executes the graph. Nil opt uses defaults. The returned error joins
// every failed step's TaskError.
func (f *Flow) Run(ctx context.Context, opt *Run) error {
	if f == nil {
		return errors.New("flow is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if opt == nil {
		opt = &Run{}
	}
	f.mu.Lock()
	if f.ran {
		f.mu.Unlock()
		return ErrAlreadyRun
	}
	f.ran = true
	f.ctx = ctx
	f.writer = opt.Writer
	if f.writer == nil {
		f.writer = os.Stdout
	}
	f.watch = opt.Watch
	f.logfn = opt.Log
	f.mu.Unlock()

	for _, t := range f.taskList() {
		t.addFallback()
	}
	f.stamp()

	c := opt.Concurrency
	if c == 0 {
		c = uint(runtime.NumCPU())
		if c == 0 {
			c = 1
		}
	}
	if n := f.subflowCount(); c <= n {
		c = n + 1
	}
	var opts []gtf.Option
	if opt.Profile != nil {
		opts = append(opts, gtf.WithProfiler())
	}
	if opt.Trace != nil {
		opts = append(opts, gtf.WithTracer())
	}
	exec := gtf.NewExecutor(c, opts...)
	exec.Run(f.tf).Wait()

	errs := f.taskErrors()
	if opt.Profile != nil {
		if err := exec.Profile(opt.Profile); err != nil {
			errs = append(errs, fmt.Errorf("profile: %w", err))
		}
	}
	if opt.Trace != nil {
		if err := exec.Trace(opt.Trace); err != nil {
			errs = append(errs, fmt.Errorf("trace: %w", err))
		}
	}
	return errors.Join(errs...)
}

func (f *Flow) taskErrors() []error {
	var errs []error
	for _, t := range f.taskList() {
		if err := t.Err(); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// Dump writes the graph in DOT format.
func (f *Flow) Dump(w io.Writer) error {
	if f == nil || f.tf == nil {
		return errors.New("flow is nil")
	}
	if w == nil {
		w = os.Stdout
	}
	return f.tf.Dump(w)
}

// List returns containers stamped with this flow's name, including stopped ones.
func (f *Flow) List(ctx context.Context) ([]response.ContainerSummary, error) {
	if f == nil {
		return nil, errors.New("flow is nil")
	}
	return List(ctx, f.cli, f.name)
}

// Remove downs every Compose project the graph referenced, then force-removes
// Container-step IDs and any leftover containers stamped with this flow's name.
func (f *Flow) Remove(ctx context.Context) error {
	if f == nil {
		return errors.New("flow is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	var errs []error
	for _, app := range f.composeList() {
		if err := app.Down(ctx, &kit.Down{RemoveVolumes: true, Writer: f.out()}); err != nil {
			errs = append(errs, err)
		}
	}
	f.mu.Lock()
	ids := append([]string(nil), f.created...)
	f.mu.Unlock()
	if f.cli == nil {
		if len(ids) > 0 {
			errs = append(errs, ErrNoClient)
		}
		return errors.Join(errs...)
	}
	for _, id := range ids {
		if err := f.cli.ContainerRemove(ctx, id, &client.Remove{Force: true}); err != nil && !client.IsErrNotFound(err) {
			errs = append(errs, err)
		}
	}
	leftover, err := f.List(ctx)
	if err != nil {
		errs = append(errs, err)
		return errors.Join(errs...)
	}
	for _, ctr := range leftover {
		if err := f.cli.ContainerRemove(ctx, ctr.ID, &client.Remove{Force: true}); err != nil && !client.IsErrNotFound(err) {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
