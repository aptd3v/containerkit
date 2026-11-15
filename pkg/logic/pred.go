package logic

// Predicate is a condition evaluated when a setter is applied, not when it is
// constructed. Use it for flags, environment, and other ambient checks.
type Predicate func() bool

// Check is a fallible predicate. Unlike [Predicate], a non-nil error stops
// apply rather than skipping the setter.
type Check func() (bool, error)

// AndFn is true when every predicate is true. Nil predicates are skipped.
func AndFn(preds ...Predicate) Predicate {
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

// And is true when every value is true. The result is a [Predicate] so it can
// be passed to [WhenFn].
func And(preds ...bool) Predicate {
	return func() bool {
		for _, pred := range preds {
			if !pred {
				return false
			}
		}
		return true
	}
}

// OrFn is true when any predicate is true. Nil predicates are skipped.
func OrFn(preds ...Predicate) Predicate {
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

// Or is true when any value is true. The result is a [Predicate] so it can be
// passed to [WhenFn].
func Or(preds ...bool) Predicate {
	return func() bool {
		for _, pred := range preds {
			if pred {
				return true
			}
		}
		return false
	}
}

func not(pred Predicate) Predicate {
	return func() bool {
		if pred == nil {
			return true
		}
		return !pred()
	}
}
