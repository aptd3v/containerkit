package kit

import (
	"errors"
	"fmt"
)

var errLogicSkip = errors.New("logic: option does not apply")

// Grouped is a bundle of setters produced by package logic. It implements
// every option interface so Group/When/Each can be passed anywhere a setter
// is accepted.
type Grouped struct {
	pred  func() bool
	check func() (bool, error)
	join  bool
	then  []any
	else_ []any
	cases []Grouped
}

// GroupedSet bundles setters applied together (errors joined).
func GroupedSet(opts []any) Grouped {
	return Grouped{join: true, then: opts}
}

// GroupedWhen applies opts when pred is true.
func GroupedWhen(pred func() bool, opts []any) Grouped {
	if pred == nil {
		pred = func() bool { return false }
	}
	return Grouped{pred: pred, then: opts}
}

// GroupedWhenElse applies thenOpts when pred is true and elseOpts otherwise.
func GroupedWhenElse(pred func() bool, thenOpts, elseOpts []any) Grouped {
	if pred == nil {
		pred = func() bool { return false }
	}
	return Grouped{pred: pred, then: thenOpts, else_: elseOpts}
}

// GroupedOnlyIf applies opts when check returns true and no error.
func GroupedOnlyIf(check func() (bool, error), opts []any) Grouped {
	return Grouped{check: check, then: opts}
}

// GroupedSwitch applies the first case whose predicate is true.
func GroupedSwitch(cases []Grouped) Grouped {
	return Grouped{cases: cases}
}

func (g Grouped) apply(scope string, call func(any) error) error {
	if g.check != nil {
		ok, err := g.check()
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		return applyLogicOpts(scope, g.then, call, g.join)
	}
	if len(g.cases) > 0 {
		for _, c := range g.cases {
			if c.pred != nil && !c.pred() {
				continue
			}
			return c.apply(scope, call)
		}
		return nil
	}
	if g.pred != nil && !g.pred() {
		return applyLogicOpts(scope, g.else_, call, g.join)
	}
	return applyLogicOpts(scope, g.then, call, g.join)
}

func applyLogicOpts(scope string, opts []any, call func(any) error, join bool) error {
	var errs []error
	for i, o := range opts {
		if o == nil {
			continue
		}
		err := call(o)
		if errors.Is(err, errLogicSkip) {
			err = cfgErr("logic", fmt.Sprintf("option[%d]", i), "option", fmt.Errorf("%w: does not apply to %s", errInvalid, scope))
		}
		if err == nil {
			continue
		}
		if !join {
			return err
		}
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}
