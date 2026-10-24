package taskflow

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/aptd3v/containerkit/pkg/kit"
	"github.com/stretchr/testify/require"
)

func TestRunTaskGraph(t *testing.T) {
	var hello atomic.Bool
	flow := New("demo", nil)
	a := flow.Task("hello", func(context.Context) error {
		hello.Store(true)
		return nil
	})
	require.NoError(t, flow.Run(context.Background(), nil))
	require.True(t, hello.Load())
	require.NoError(t, a.Err())
	require.ErrorIs(t, flow.Run(context.Background(), nil), ErrAlreadyRun)
}

func TestFailedStepSkipsDependents(t *testing.T) {
	var ranB, ranC atomic.Bool
	flow := New("skip", nil)
	a := flow.Task("a", func(context.Context) error {
		return errors.New("boom")
	})
	b := flow.Task("b", func(context.Context) error {
		ranB.Store(true)
		return nil
	})
	c := flow.Task("c", func(context.Context) error {
		ranC.Store(true)
		return nil
	})
	a.Precede(b)
	err := flow.Run(context.Background(), nil)
	require.Error(t, err)
	var te *TaskError
	require.ErrorAs(t, err, &te)
	require.Equal(t, "a", te.Task)
	require.False(t, ranB.Load())
	require.True(t, ranC.Load())
	require.Nil(t, b.Err())
	require.Nil(t, c.Err())
}

func TestSiblingsRunWhenOneFails(t *testing.T) {
	var ranBundle atomic.Bool
	var ranPublish atomic.Bool
	flow := New("fan", nil)
	check := flow.Task("check", func(context.Context) error {
		return errors.New("bad report")
	})
	bundle := flow.Task("bundle", func(context.Context) error {
		ranBundle.Store(true)
		return nil
	})
	publish := flow.Task("publish", func(context.Context) error {
		ranPublish.Store(true)
		return nil
	})
	check.Precede(publish)
	bundle.Precede(publish)
	err := flow.Run(context.Background(), nil)
	require.Error(t, err)
	require.True(t, ranBundle.Load())
	require.False(t, ranPublish.Load())
}

func TestConditionSelectsBranch(t *testing.T) {
	var ran0, ran1 atomic.Bool
	flow := New("cond", nil)
	start := flow.Task("start", func(context.Context) error { return nil })
	gate := flow.Condition("gate", func(context.Context) uint { return 1 })
	b0 := flow.Task("b0", func(context.Context) error {
		ran0.Store(true)
		return nil
	})
	b1 := flow.Task("b1", func(context.Context) error {
		ran1.Store(true)
		return nil
	})
	start.Precede(gate)
	gate.Precede(b0, b1)
	require.NoError(t, flow.Run(context.Background(), nil))
	require.False(t, ran0.Load())
	require.True(t, ran1.Load())
}

func TestConditionRetryLoop(t *testing.T) {
	var n atomic.Int32
	flow := New("retry", nil)
	start := flow.Task("start", func(context.Context) error { return nil })
	integ := flow.Task("integ", func(context.Context) error {
		v := n.Add(1)
		if v < 3 {
			return errors.New("flaky")
		}
		return nil
	})
	verify := flow.Condition("verify", func(context.Context) uint {
		if integ.Err() == nil || n.Load() >= 3 {
			return 0
		}
		return 1
	})
	done := flow.Task("done", func(context.Context) error { return nil })
	start.Precede(integ)
	integ.Precede(verify)
	verify.Precede(done, integ)
	require.NoError(t, flow.Run(context.Background(), nil))
	require.Equal(t, int32(3), n.Load())
	require.NoError(t, done.Err())
}

func TestSubflowRunsInnerTask(t *testing.T) {
	var inner atomic.Bool
	flow := New("sf", nil)
	flow.Subflow("tests", func(sf *Subflow) {
		sf.Task("unit", func(context.Context) error {
			inner.Store(true)
			return nil
		}).Priority(PriorityHigh)
	})
	require.NoError(t, flow.Run(context.Background(), &Run{Concurrency: 2}))
	require.True(t, inner.Load())
}

func TestSubflowCycleRetries(t *testing.T) {
	var n atomic.Int32
	flow := New("sf-cycle", nil)
	flow.Subflow("loop", func(sf *Subflow) {
		enter := sf.Task("enter", func(context.Context) error { return nil })
		body := sf.Task("body", func(context.Context) error {
			n.Add(1)
			return nil
		})
		verify := sf.Condition("verify", func(context.Context) uint {
			if n.Load() >= 3 {
				return 0
			}
			return 1
		})
		done := sf.Task("done", func(context.Context) error { return nil })
		enter.Precede(body)
		body.Precede(verify)
		verify.Precede(done, body)
	})
	require.NoError(t, flow.Run(context.Background(), &Run{Concurrency: 2}))
	require.Equal(t, int32(3), n.Load())
}

func TestWatchEvents(t *testing.T) {
	var mu sync.Mutex
	var kinds []EventKind
	flow := New("watch", nil)
	a := flow.Task("a", func(context.Context) error { return nil })
	b := flow.Task("b", func(context.Context) error { return errors.New("nope") })
	c := flow.Task("c", func(context.Context) error { return nil })
	a.Precede(b)
	b.Precede(c)
	err := flow.Run(context.Background(), &Run{
		Watch: func(e Event) {
			if e.Name != "b" && e.Name != "c" {
				return
			}
			mu.Lock()
			kinds = append(kinds, e.Kind)
			mu.Unlock()
		},
	})
	require.Error(t, err)
	require.Error(t, b.Err())
	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, []EventKind{EventStart, EventError, EventSkip}, kinds)
}

func TestEmitLog(t *testing.T) {
	var mu sync.Mutex
	var got string
	flow := New("log", nil)
	flow.logfn = func(step string, p []byte) {
		mu.Lock()
		got += step + ":" + string(p)
		mu.Unlock()
	}
	n, err := (logWriter{f: flow, step: "build"}).Write([]byte("layer"))
	require.NoError(t, err)
	require.Equal(t, 5, n)
	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, "build:layer", got)
}

func TestDumpContainsTaskNames(t *testing.T) {
	flow := New("dot", nil)
	a := flow.Task("build", func(context.Context) error { return nil })
	b := flow.Task("test", func(context.Context) error { return nil })
	a.Precede(b)
	var buf bytes.Buffer
	require.NoError(t, flow.Dump(&buf))
	out := buf.String()
	require.Contains(t, out, "build")
	require.Contains(t, out, "test")
}

func TestProfileAndTrace(t *testing.T) {
	var prof, trace bytes.Buffer
	flow := New("obs", nil)
	flow.Task("a", func(context.Context) error { return nil })
	require.NoError(t, flow.Run(context.Background(), &Run{Profile: &prof, Trace: &trace}))
	require.NotEmpty(t, prof.String())
	require.NotEmpty(t, trace.String())
}

func TestAlwaysRunsAfterFailureAndStillBlocks(t *testing.T) {
	var ranDown, ranAfter atomic.Bool
	flow := New("always", nil)
	up := flow.Task("up", func(context.Context) error { return errors.New("boom") })
	test := flow.Task("test", func(context.Context) error { return nil })
	down := flow.Task("down", func(context.Context) error {
		ranDown.Store(true)
		return nil
	}).Always()
	after := flow.Task("after", func(context.Context) error {
		ranAfter.Store(true)
		return nil
	})
	up.Precede(test)
	test.Precede(down)
	down.Precede(after)
	err := flow.Run(context.Background(), nil)
	require.Error(t, err)
	require.True(t, ranDown.Load())
	require.NoError(t, down.Err())
	require.False(t, ranAfter.Load())
}

func TestSubflowFailureSkipsSuccessors(t *testing.T) {
	var ranAfter atomic.Bool
	flow := New("sf-fail", nil)
	tests := flow.Subflow("tests", func(sf *Subflow) {
		sf.Task("unit", func(context.Context) error { return errors.New("red") })
		sf.Task("lint", func(context.Context) error { return nil })
	})
	after := flow.Task("after", func(context.Context) error {
		ranAfter.Store(true)
		return nil
	})
	tests.Precede(after)
	err := flow.Run(context.Background(), &Run{Concurrency: 2})
	require.Error(t, err)
	var te *TaskError
	require.ErrorAs(t, err, &te)
	require.Equal(t, "unit", te.Task)
	require.False(t, ranAfter.Load())
	require.NoError(t, tests.Err())
}

func TestSkippedSubflowDoesNotHang(t *testing.T) {
	var inner atomic.Bool
	flow := New("sf-skip", nil)
	a := flow.Task("a", func(context.Context) error { return errors.New("boom") })
	sf := flow.Subflow("tests", func(sf *Subflow) {
		sf.Task("unit", func(context.Context) error {
			inner.Store(true)
			return nil
		})
	})
	a.Precede(sf)
	require.Error(t, flow.Run(context.Background(), &Run{Concurrency: 2}))
	require.False(t, inner.Load())
}

func TestRetrySucceeds(t *testing.T) {
	var n atomic.Int32
	var mu sync.Mutex
	var retries []int
	flow := New("retry", nil)
	flaky := flow.Task("flaky", func(context.Context) error {
		if n.Add(1) < 3 {
			return errors.New("flaky")
		}
		return nil
	}).Retry(2)
	require.NoError(t, flow.Run(context.Background(), &Run{Watch: func(e Event) {
		if e.Kind == EventRetry {
			require.Same(t, flaky, e.Task)
			require.Error(t, e.Err)
			mu.Lock()
			retries = append(retries, e.Attempt)
			mu.Unlock()
		}
	}}))
	require.Equal(t, int32(3), n.Load())
	require.NoError(t, flaky.Err())
	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, []int{1, 2}, retries)
}

func TestRetryExhausted(t *testing.T) {
	var n atomic.Int32
	flow := New("retry-fail", nil)
	flow.Task("flaky", func(context.Context) error {
		n.Add(1)
		return errors.New("still flaky")
	}).Retry(1)
	err := flow.Run(context.Background(), nil)
	require.Error(t, err)
	require.Equal(t, int32(2), n.Load())
}

func TestIfThenElse(t *testing.T) {
	for _, want := range []bool{true, false} {
		var ranThen, ranElse atomic.Bool
		flow := New("if", nil)
		gate := flow.If("gate", func(context.Context) bool { return want })
		then := flow.Task("then", func(context.Context) error {
			ranThen.Store(true)
			return nil
		})
		els := flow.Task("else", func(context.Context) error {
			ranElse.Store(true)
			return nil
		})
		gate.Then(then).Else(els)
		require.NoError(t, flow.Run(context.Background(), nil))
		require.Equal(t, want, ranThen.Load())
		require.Equal(t, !want, ranElse.Load())
	}
}

func TestIfWithoutElse(t *testing.T) {
	var ranThen atomic.Bool
	flow := New("if-noelse", nil)
	gate := flow.If("gate", func(context.Context) bool { return false })
	gate.Then(flow.Task("then", func(context.Context) error {
		ranThen.Store(true)
		return nil
	}))
	require.NoError(t, flow.Run(context.Background(), nil))
	require.False(t, ranThen.Load())
}

func TestIfSkipsWhenPredecessorFails(t *testing.T) {
	var ranThen, ranElse, predicate atomic.Bool
	flow := New("if-skip", nil)
	a := flow.Task("a", func(context.Context) error { return errors.New("boom") })
	gate := flow.If("gate", func(context.Context) bool {
		predicate.Store(true)
		return true
	})
	then := flow.Task("then", func(context.Context) error {
		ranThen.Store(true)
		return nil
	})
	els := flow.Task("else", func(context.Context) error {
		ranElse.Store(true)
		return nil
	})
	a.Precede(gate)
	gate.Then(then).Else(els)
	require.Error(t, flow.Run(context.Background(), nil))
	require.False(t, predicate.Load())
	require.False(t, ranThen.Load())
	require.False(t, ranElse.Load())
}

func TestConditionOutOfRangeFailsStep(t *testing.T) {
	var ran atomic.Bool
	flow := New("cond-range", nil)
	gate := flow.Condition("gate", func(context.Context) uint { return 7 })
	gate.Precede(flow.Task("only", func(context.Context) error {
		ran.Store(true)
		return nil
	}))
	err := flow.Run(context.Background(), nil)
	require.Error(t, err)
	var te *TaskError
	require.ErrorAs(t, err, &te)
	require.Equal(t, "gate", te.Task)
	require.False(t, ran.Load())
}

func TestConditionWithoutSuccessors(t *testing.T) {
	flow := New("cond-none", nil)
	flow.Condition("gate", func(context.Context) uint { return 0 })
	err := flow.Run(context.Background(), nil)
	require.Error(t, err)
	var te *TaskError
	require.ErrorAs(t, err, &te)
	require.Equal(t, "gate", te.Task)
}

func TestConditionPanicFailsStep(t *testing.T) {
	flow := New("cond-panic", nil)
	gate := flow.Condition("gate", func(context.Context) uint { panic("nope") })
	gate.Precede(flow.Task("only", func(context.Context) error { return nil }))
	err := flow.Run(context.Background(), nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "panic")
}

func TestDependsOn(t *testing.T) {
	var order []string
	var mu sync.Mutex
	rec := func(name string) func(context.Context) error {
		return func(context.Context) error {
			mu.Lock()
			order = append(order, name)
			mu.Unlock()
			return nil
		}
	}
	flow := New("deps", nil)
	a := flow.Task("a", rec("a"))
	b := flow.Task("b", rec("b"))
	b.DependsOn(a)
	require.NoError(t, flow.Run(context.Background(), &Run{Concurrency: 1}))
	require.Equal(t, []string{"a", "b"}, order)
}

func TestFlowNameAndList(t *testing.T) {
	flow := New("named", nil)
	require.Equal(t, "named", flow.Name())
	_, err := flow.List(context.Background())
	require.ErrorIs(t, err, ErrNoClient)
}

func TestTrackComposeDedup(t *testing.T) {
	flow := New("ci", nil)
	db := kit.NewCompose(kit.NewProject("db"))
	other := kit.NewCompose(kit.NewProject("other"))
	flow.Compose("up", db, &kit.Up{})
	flow.Compose("down", db, &kit.Down{})
	flow.Compose("x", other, func(context.Context, *kit.Compose) error { return nil })
	flow.Compose("nil", nil, func(context.Context, *kit.Compose) error { return nil })
	require.Equal(t, []*kit.Compose{db, other}, flow.composeList())
}

func TestRemoveWithoutClientOrWork(t *testing.T) {
	flow := New("empty", nil)
	require.NoError(t, flow.Remove(context.Background()))
}

func TestComposeUpDownRequireApp(t *testing.T) {
	flow := New("cmp-nil", nil)
	up := flow.Compose("up", nil, &kit.Up{})
	down := flow.Compose("down", nil, &kit.Down{}).Always()
	up.Precede(down)
	err := flow.Run(context.Background(), nil)
	require.Error(t, err)
	require.ErrorIs(t, up.Err(), errNilCompose)
	require.ErrorIs(t, down.Err(), errNilCompose)
}

func TestEventCarriesTask(t *testing.T) {
	var mu sync.Mutex
	var seen []*Task
	flow := New("evt", nil)
	a := flow.Task("a", func(context.Context) error { return nil })
	require.NoError(t, flow.Run(context.Background(), &Run{Watch: func(e Event) {
		mu.Lock()
		seen = append(seen, e.Task)
		mu.Unlock()
	}}))
	mu.Lock()
	defer mu.Unlock()
	require.NotEmpty(t, seen)
	for _, s := range seen {
		require.Same(t, a, s)
	}
}

func TestPanicIsTaskError(t *testing.T) {
	var sibling atomic.Bool
	flow := New("panic", nil)
	flow.Task("boom", func(context.Context) error {
		panic("nope")
	})
	flow.Task("ok", func(context.Context) error {
		sibling.Store(true)
		return nil
	})
	err := flow.Run(context.Background(), nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "panic")
	require.True(t, sibling.Load())
}

func TestComposeStepCallsFunc(t *testing.T) {
	var called atomic.Bool
	flow := New("cmp", nil)
	flow.Compose("up", nil, func(ctx context.Context, app *kit.Compose) error {
		require.Nil(t, app)
		called.Store(true)
		return nil
	})
	require.NoError(t, flow.Run(context.Background(), nil))
	require.True(t, called.Load())
}

func TestBuildAndContainerRequireClient(t *testing.T) {
	t.Run("container", func(t *testing.T) {
		flow := New("ctr", nil)
		flow.Container("c", kit.NewContainer("c").Image("alpine"), nil)
		err := flow.Run(context.Background(), nil)
		require.ErrorIs(t, err, ErrNoClient)
	})
	t.Run("build", func(t *testing.T) {
		flow := New("bld", nil)
		flow.Build("b", kit.NewBuild().Dockerfile(kit.NewDockerFile().From("alpine:latest")), nil)
		err := flow.Run(context.Background(), nil)
		require.ErrorIs(t, err, ErrNoClient)
	})
}

func TestCanceledContextFailsStep(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	flow := New("cancel", nil)
	flow.Task("a", func(context.Context) error { return nil })
	err := flow.Run(ctx, nil)
	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
}
