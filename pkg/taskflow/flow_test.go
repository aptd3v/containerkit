package taskflow

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/aptd3v/containerkit/pkg/kit"
)

func TestNilReceivers(t *testing.T) {
	var f *Flow
	var task *Task
	var sf Subflow
	if f.Name() != "" || f.Task("a", nil) != nil || f.Subflow("a", nil) != nil || f.Condition("a", nil) != nil || f.If("a", nil) != nil {
		t.Fatal("nil flow constructors")
	}
	if f.Container("a", nil, nil) != nil || f.Build("a") != nil || f.Compose("a", nil, nil) != nil {
		t.Fatal("nil flow steps")
	}
	if task.Precede(nil) != nil || task.DependsOn(nil) != nil || task.Then(nil) != nil || task.Else(nil) != nil {
		t.Fatal("nil task edges")
	}
	if task.Priority(PriorityHigh) != nil || task.Always() != nil || task.Retry(1) != nil {
		t.Fatal("nil task props")
	}
	if task.Name() != "" || task.ID() != "" || task.ExitCode() != 0 || task.Err() != nil {
		t.Fatal("nil task readers")
	}
	if sf.Task("a", nil) != nil || sf.Subflow("a", nil) != nil || sf.Condition("a", nil) != nil || sf.If("a", nil) != nil {
		t.Fatal("nil subflow")
	}
	if sf.Container("a", nil, nil) != nil || sf.Build("a") != nil || sf.Compose("a", nil, nil) != nil {
		t.Fatal("nil subflow steps")
	}
	if err := f.Run(context.Background(), nil); err == nil {
		t.Fatal("nil run")
	}
	if _, err := f.List(context.Background()); err == nil {
		t.Fatal("nil list")
	}
	if err := f.Remove(context.Background()); err == nil || f.Dump(&bytes.Buffer{}) == nil {
		t.Fatal("nil remove/dump")
	}
	if f.context() == nil || f.engine() != nil || f.composer() != nil || f.composeList() != nil || f.taskList() != nil {
		t.Fatal("nil helpers")
	}
	f.stamp()
	f.emit(Event{})
	f.emitLog("a", []byte("x"))
	f.trackCompose(nil)
	logWriter{}.Write(nil)
	logWriter{}.Write([]byte("x"))
}

func TestTaskGraph(t *testing.T) {
	ctx := context.Background()
	f := New("", nil)
	if f.Name() != "taskflow" {
		t.Fatalf("name %q", f.Name())
	}
	var ran bool
	a := f.Task("a", func(context.Context) error {
		ran = true
		return nil
	})
	b := f.Task("b", nil)
	a.Precede(nil, b)
	b.DependsOn(nil)
	a.Then(b)
	a.Else(b)
	a.Priority(0)
	a.Priority(PriorityHigh)
	b.Priority(PriorityNormal)
	b.Retry(-1)
	if err := f.Run(ctx, discardRun()); err != nil {
		t.Fatal(err)
	}
	if !ran || b.Err() != nil || a.Name() != "a" {
		t.Fatalf("ran %v err %v", ran, b.Err())
	}
	if err := f.Run(ctx, discardRun()); !errors.Is(err, ErrAlreadyRun) {
		t.Fatalf("second run %v", err)
	}
}

func TestSkipAlwaysRetryPanicCancel(t *testing.T) {
	ctx := context.Background()
	f := New("g", nil)
	var events []EventKind
	var mu sync.Mutex
	watch := func(e Event) {
		mu.Lock()
		events = append(events, e.Kind)
		mu.Unlock()
	}
	bad := f.Task("bad", func(context.Context) error { return errors.New("no") })
	skipped := f.Task("skip", func(context.Context) error { return nil })
	always := f.Task("always", func(context.Context) error { return nil }).Always()
	after := f.Task("after", func(context.Context) error { return nil })
	bad.Precede(skipped, always)
	always.Precede(after)
	if err := f.Run(ctx, &Run{Writer: io.Discard, Concurrency: 1, Watch: watch}); err == nil {
		t.Fatal("expected failure")
	}
	if skipped.Err() != nil || always.Err() != nil || after.Err() != nil {
		t.Fatalf("skip %v always %v after %v", skipped.Err(), always.Err(), after.Err())
	}

	f = New("retry", nil)
	n := 0
	step := f.Task("flaky", func(context.Context) error {
		n++
		if n < 3 {
			return errors.New("flaky")
		}
		return nil
	}).Retry(2)
	if err := f.Run(ctx, discardRun()); err != nil || step.Err() != nil || n != 3 {
		t.Fatalf("retry %v n %d err %v", err, n, step.Err())
	}

	f = New("exhaust", nil)
	f.Task("nope", func(context.Context) error { return errors.New("no") }).Retry(1)
	if err := f.Run(ctx, discardRun()); err == nil {
		t.Fatal("expected exhaust")
	}

	f = New("panic", nil)
	p := f.Task("boom", func(context.Context) error { panic("boom") }).Retry(4)
	if err := f.Run(ctx, discardRun()); err == nil || p.Err() == nil {
		t.Fatal(err)
	}

	cancelCtx, cancel := context.WithCancel(ctx)
	f = New("cancel", nil)
	attempts := 0
	f.Task("c", func(context.Context) error {
		attempts++
		cancel()
		return errors.New("x")
	}).Retry(3)
	if err := f.Run(cancelCtx, discardRun()); err == nil || attempts != 1 {
		t.Fatalf("cancel %v attempts %d", err, attempts)
	}

	doneCtx, done := context.WithCancel(ctx)
	done()
	f = New("precancel", nil)
	called := false
	f.Task("c", func(context.Context) error {
		called = true
		return nil
	}).Retry(2)
	if err := f.Run(doneCtx, &Run{Writer: io.Discard}); err == nil || called {
		t.Fatalf("precancel %v called %v", err, called)
	}
	if err := f.Run(nil, discardRun()); !errors.Is(err, ErrAlreadyRun) {
		t.Fatal(err)
	}
}

func TestIfAndCondition(t *testing.T) {
	ctx := context.Background()
	f := New("if", nil)
	yesRan, noRan := false, false
	iff := f.If("gate", func(context.Context) bool { return true })
	yes := f.Task("yes", func(context.Context) error { yesRan = true; return nil })
	no := f.Task("no", func(context.Context) error { noRan = true; return nil })
	iff.Then(yes).Else(no)
	if err := f.Run(ctx, discardRun()); err != nil || !yesRan || noRan {
		t.Fatalf("then %v yes %v no %v", err, yesRan, noRan)
	}

	f = New("else", nil)
	yesRan, noRan = false, false
	iff = f.If("gate", func(context.Context) bool { return false })
	yes = f.Task("yes", func(context.Context) error { yesRan = true; return nil })
	no = f.Task("no", func(context.Context) error { noRan = true; return nil })
	iff.Else(no).Then(yes)
	iff.Precede(yes)
	if err := f.Run(ctx, discardRun()); err != nil || yesRan || !noRan {
		t.Fatalf("else %v yes %v no %v", err, yesRan, noRan)
	}

	f = New("if-skip", nil)
	yesRan, noRan = false, false
	bad := f.Task("bad", func(context.Context) error { return errors.New("no") })
	iff = f.If("gate", func(context.Context) bool { return true })
	yes = f.Task("yes", func(context.Context) error { yesRan = true; return nil })
	no = f.Task("no", func(context.Context) error { noRan = true; return nil })
	bad.Precede(iff)
	iff.Then(yes).Else(no)
	if err := f.Run(ctx, discardRun()); err == nil || yesRan || noRan {
		t.Fatalf("if skip %v yes %v no %v", err, yesRan, noRan)
	}

	f = New("if-nil", nil)
	f.If("gate", nil)
	if err := f.Run(ctx, discardRun()); err != nil {
		t.Fatal(err)
	}

	f = New("cond", nil)
	got := ""
	cond := f.Condition("pick", func(context.Context) uint { return 1 })
	a := f.Task("a", func(context.Context) error { got = "a"; return nil })
	b := f.Task("b", func(context.Context) error { got = "b"; return nil })
	cond.Precede(a, b)
	if err := f.Run(ctx, discardRun()); err != nil || got != "b" {
		t.Fatalf("cond %v got %s", err, got)
	}

	f = New("cond-range", nil)
	cond = f.Condition("pick", func(context.Context) uint { return 9 })
	ran := false
	next := f.Task("next", func(context.Context) error { ran = true; return nil })
	cond.Precede(next)
	if err := f.Run(ctx, discardRun()); err == nil || ran || cond.Err() == nil {
		t.Fatalf("range %v ran %v", err, ran)
	}

	f = New("cond-after", nil)
	bad = f.Task("bad", func(context.Context) error { return errors.New("no") })
	got = ""
	cond = f.Condition("pick", func(context.Context) uint { return 0 })
	a = f.Task("a", func(context.Context) error { got = "a"; return nil })
	bad.Precede(cond)
	cond.Precede(a)
	if err := f.Run(ctx, discardRun()); err == nil || got != "a" {
		t.Fatalf("cond after %v got %s", err, got)
	}

	f = New("cond-nil", nil)
	f.Condition("pick", nil)
	if err := f.Run(ctx, discardRun()); err != nil {
		t.Fatal(err)
	}
}

func TestSubflow(t *testing.T) {
	ctx := context.Background()
	f := New("sub", nil)
	innerRan := false
	outerRan := false
	sub := f.Subflow("sf", func(sf *Subflow) {
		inner := sf.Task("inner", func(context.Context) error {
			innerRan = true
			return errors.New("inner")
		})
		c := sf.Condition("c", func(context.Context) uint { return 0 })
		c.Precede(inner)
		sf.If("g", func(context.Context) bool { return true })
		sf.Subflow("nested", func(sf *Subflow) {
			sf.Task("deep", func(context.Context) error { return nil })
		})
		sf.Build("b")
		sf.Container("ctr", nil, nil)
		sf.Compose("cmp", nil, nil)
	})
	outer := f.Task("outer", func(context.Context) error { outerRan = true; return nil })
	sub.Precede(outer)
	f.Subflow("empty", nil)
	err := f.Run(ctx, &Run{Writer: io.Discard, Concurrency: 1})
	if err == nil || !innerRan || outerRan {
		t.Fatalf("sub %v inner %v outer %v", err, innerRan, outerRan)
	}
}

func TestDumpProfileTrace(t *testing.T) {
	f := New("dot", nil)
	f.Task("a", func(context.Context) error { return nil }).Priority(PriorityLow)
	var dot bytes.Buffer
	if err := f.Dump(&dot); err != nil || !bytes.Contains(dot.Bytes(), []byte("digraph")) {
		t.Fatalf("dump %v %s", err, dot.String())
	}
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	err = f.Dump(nil)
	w.Close()
	os.Stdout = old
	slurp, _ := io.ReadAll(r)
	r.Close()
	if err != nil || !bytes.Contains(slurp, []byte("digraph")) {
		t.Fatalf("stdout dump %v %s", err, slurp)
	}

	var prof, trace bytes.Buffer
	if err := f.Run(nil, &Run{Writer: io.Discard, Profile: &prof, Trace: &trace}); err != nil {
		t.Fatal(err)
	}
	if prof.Len() == 0 || trace.Len() == 0 {
		t.Fatalf("profile %q trace %q", prof.String(), trace.String())
	}

	f = New("failio", nil)
	f.Task("a", func(context.Context) error { return nil })
	err = f.Run(context.Background(), &Run{Writer: io.Discard, Profile: failWriter{}, Trace: failWriter{}})
	if err == nil || !bytes.Contains([]byte(err.Error()), []byte("profile")) {
		t.Fatal(err)
	}
}

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errors.New("full") }

func TestCycle(t *testing.T) {
	ctx := context.Background()
	f := New("loop", nil)
	ran := 0
	a := f.Task("a", func(context.Context) error {
		ran++
		return nil
	})
	b := f.Task("b", func(context.Context) error {
		ran++
		return nil
	})
	a.DependsOn(b)
	b.DependsOn(a)
	err := f.Run(ctx, discardRun())
	if err == nil || !cycleNames(err, "a", "b") || ran != 0 {
		t.Fatalf("run %v ran %d", err, ran)
	}

	f = New("sub", nil)
	f.Subflow("outer", func(sf *Subflow) {
		left := sf.Task("left", func(context.Context) error {
			ran++
			return nil
		})
		right := sf.Task("right", func(context.Context) error {
			ran++
			return nil
		})
		left.DependsOn(right)
		right.DependsOn(left)
	})
	err = f.Run(ctx, discardRun())
	if err == nil || !cycleNames(err, "left", "right") || ran != 0 {
		t.Fatalf("subflow %v ran %d", err, ran)
	}

	f = New("out", nil)
	f.Subflow("sf", func(sf *Subflow) {
		inner := sf.Task("inner", func(context.Context) error { return nil })
		inner.succs = append(inner.succs, nil)
	})
	if err = f.Run(ctx, discardRun()); err != nil {
		t.Fatal(err)
	}

	prev := runIDEntropy
	runIDEntropy = bytes.NewReader(nil)
	t.Cleanup(func() { runIDEntropy = prev })
	f = New("id", nil)
	f.Task("a", func(context.Context) error { return nil })
	if err := f.Run(ctx, discardRun()); err == nil {
		t.Fatal("run id")
	}
}

func cycleNames(err error, a, b string) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, a+" -> "+b+" -> "+a) || strings.Contains(msg, b+" -> "+a+" -> "+b)
}

func TestLoggerEvents(t *testing.T) {
	var got []kit.LogEvent
	var mu sync.Mutex
	log := kit.NewFrameLogger(func(e kit.LogEvent) {
		mu.Lock()
		got = append(got, e)
		mu.Unlock()
	})
	f := New("log", nil)
	f.Task("a", func(context.Context) error { return errors.New("no") })
	ctx := log.Context(context.Background())
	if err := f.Run(ctx, discardRun()); err == nil {
		t.Fatal("expected error")
	}
	if len(got) == 0 {
		t.Fatal("no events")
	}
}
