package containerkit

import "errors"

// Mutator mutates a container. Used with Apply / WhenTrue / Group.
type Mutator func(*Container)

// Apply applies mutators to the container and returns it.
func (c *Container) Apply(fns ...Mutator) *Container {
	for _, fn := range fns {
		if fn != nil {
			fn(c)
		}
	}
	return c
}

type PredicateClosure func() bool
type CheckClosure func() (bool, error)

func WhenTrueFn(predicate PredicateClosure, fns ...Mutator) Mutator {
	return func(c *Container) {
		if predicate == nil || !predicate() {
			return
		}
		for _, fn := range fns {
			if fn != nil {
				fn(c)
			}
		}
	}
}

func WhenTrue(check bool, fns ...Mutator) Mutator {
	return WhenTrueFn(func() bool { return check }, fns...)
}

func WhenTrueElseFn(predicate PredicateClosure, fn, elseFn Mutator) Mutator {
	return func(c *Container) {
		if predicate != nil && predicate() {
			if fn != nil {
				fn(c)
			}
			return
		}
		if elseFn != nil {
			elseFn(c)
		}
	}
}

func WhenTrueElse(check bool, fn, elseFn Mutator) Mutator {
	return WhenTrueElseFn(func() bool { return check }, fn, elseFn)
}

func AndFn(preds ...PredicateClosure) func() bool {
	return func() bool {
		for _, pred := range preds {
			if pred == nil {
				continue
			}
			if !pred() {
				return false
			}
		}
		return true
	}
}

func And(preds ...bool) func() bool {
	return func() bool {
		for _, pred := range preds {
			if !pred {
				return false
			}
		}
		return true
	}
}

func OrFn(preds ...PredicateClosure) func() bool {
	return func() bool {
		for _, pred := range preds {
			if pred == nil {
				continue
			}
			if pred() {
				return true
			}
		}
		return false
	}
}

func Or(preds ...bool) func() bool {
	return func() bool {
		for _, pred := range preds {
			if pred {
				return true
			}
		}
		return false
	}
}

func Group(fns ...Mutator) Mutator {
	return func(c *Container) {
		for _, fn := range fns {
			if fn != nil {
				fn(c)
			}
		}
	}
}

func OnlyIf(check CheckClosure, f Mutator) Mutator {
	return func(c *Container) {
		if check == nil {
			c.Errors = append(c.Errors, errors.New("containerkit.OnlyIf: check closure is nil"))
			return
		}
		ok, err := check()
		if err != nil {
			c.Errors = append(c.Errors, err)
			return
		}
		if f != nil && ok {
			f(c)
		}
	}
}

func Each[T any](items []T, f func(i int, t T) Mutator) Mutator {
	return func(c *Container) {
		if len(items) == 0 || f == nil {
			return
		}
		for i, item := range items {
			if fn := f(i, item); fn != nil {
				fn(c)
			}
		}
	}
}

// WhenTrueOpt applies option funcs (SetContainerConfig, SetServiceConfig, etc.) only if check is true.
func WhenTrueOpt[T any, O ~func(T) error](check bool, fns ...O) O {
	return func(t T) error {
		if !check {
			return nil
		}
		for _, fn := range fns {
			if fn != nil {
				if err := fn(t); err != nil {
					return err
				}
			}
		}
		return nil
	}
}
