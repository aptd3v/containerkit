// Package tools provides various helpers for writing declarative option setters.
package tools

// PredicateClosure is a condition based on external or ambient context,
// not the internal config. This allows behavior to be toggled based on
// CLI flags, environment variables, testing hooks, etc.
type PredicateClosure func() bool

// WhenTrue is a function that takes a variadic number of setters and returns a single setter.
// Only if it passes the provided predicate closures, the setters will be called.
//
// note: if any of the setters are nil, they will be skipped and not added to warnings
func WhenTrue[T any, O ~func(T) error](predicate PredicateClosure, fns ...O) O {
	return func(t T) error {
		if !predicate() {
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
func WhenTrueElse[T any, O ~func(T) error](predicate PredicateClosure, fns O, elseFn O) O {
	return func(t T) error {
		if predicate() {
			return fns(t)
		}
		return elseFn(t)
	}
}

// And is a function that takes a variadic number of predicates and returns a single predicate.
// It returns true if all the predicates are true.
func And(preds ...PredicateClosure) func() bool {
	return func() bool {
		for _, pred := range preds {
			if !pred() {
				return false
			}
		}
		return true
	}
}

// Or is a function that takes a variadic number of predicates and returns a single predicate.
//
// It returns true if any of the predicates are true.
func Or(preds ...PredicateClosure) func() bool {
	return func() bool {
		for _, pred := range preds {
			if pred() {
				return true
			}
		}
		return false
	}
}

// Group combines multiple setters into a single setter
//
// note: if any of the setters are nil, they will be skipped and not added to warnings
func Group[T any, O ~func(T) error](fns ...O) O {
	return func(t T) error {

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

func AlwaysTrue() PredicateClosure {
	return func() bool {
		return true
	}
}
