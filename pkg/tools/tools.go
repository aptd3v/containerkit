// Package tools provides various helpers for writing declarative option setters.
package tools

// PredicateClosure is a condition based on external or ambient context,
// not the internal config. This allows behavior to be toggled based on
// CLI flags, environment variables, testing hooks, etc.
type PredicateClosure func() bool

// WhenTrueFn is a function that takes a variadic number of setters and returns a single setter.
// Only if it passes the provided predicate closures, the setters will be called.
//
// note: if any of the setters are nil, they will be skipped and not added to warnings
func WhenTrueFn[T any, O ~func(T) error](predicate PredicateClosure, fns ...O) O {
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

// WhenTrue is a function that takes a boolean and a variadic number of setters and returns a single setter.
// Only if the boolean is true, the setters will be called.
//
// note: if any of the setters are nil, they will be skipped and not added to warnings
func WhenTrue[T any, O ~func(T) error](check bool, fns ...O) O {
	return WhenTrueFn(func() bool { return check }, fns...)
}

// WhenTrueElseFn is a function that takes a predicate closure, a function to call if the predicate is true,
// and a function to call if the predicate is false.
// It returns a single setter.
// Only if it passes the provided predicate closure, the setters will be called.
//
// note: if any of the setters are nil, they will be skipped and not added to warnings
func WhenTrueElseFn[T any, O ~func(T) error](predicate PredicateClosure, fns O, elseFn O) O {
	return func(t T) error {
		if predicate() {
			if fns != nil {
				return fns(t)
			}
			return nil
		}
		if elseFn != nil {
			return elseFn(t)
		}
		return nil
	}
}

// WhenTrueElse is a function that takes a boolean, a function to call if the boolean is true,
// and a function to call if the boolean is false.
// It returns a single setter.
// Only if it passes the provided predicate closure, the setters will be called.
//
// note: if any of the setters are nil, they will be skipped and not added to warnings
func WhenTrueElse[T any, O ~func(T) error](check bool, fns O, elseFn O) O {
	return WhenTrueElseFn(func() bool { return check }, fns, elseFn)
}

// AndFn is a function that takes a variadic number of predicates and returns a single predicate.
// It returns true if all the predicates are true.
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

// And is a function that takes a variadic number of booleans and returns true if all the booleans are true.
func And(preds ...bool) func() bool {
	return AndFn(func() bool {
		for _, pred := range preds {
			if !pred {
				return false
			}
		}
		return true
	})
}

// OrFn is a function that takes a variadic number of predicates and returns a single predicate.
//
// It returns true if any of the predicates are true.
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

// Or is a function that takes a variadic number of booleans and returns true if any of the booleans are true.
func Or(preds ...bool) func() bool {
	return OrFn(func() bool {
		for _, pred := range preds {
			if pred {
				return true
			}
		}
		return false
	})
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
