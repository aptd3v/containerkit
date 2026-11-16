package logic

import (
	"errors"

	"github.com/aptd3v/containerkit/pkg/kit"
)

var errNilCheck = errors.New("logic.onlyif: check is nil")

// Group combines setters into one. Errors from the inner setters are joined.
// Nil setters are skipped.
func Group(opts ...any) kit.Grouped {
	return kit.GroupedSet(opts)
}

// When applies setters when cond is true. Nil setters are skipped.
func When(cond bool, opts ...any) kit.Grouped {
	return WhenFn(func() bool { return cond }, opts...)
}

// WhenFn applies setters when pred() is true. A nil predicate skips the setters.
func WhenFn(pred Predicate, opts ...any) kit.Grouped {
	return kit.GroupedWhen(pred, opts)
}

// WhenElse applies thenOpt when cond is true and elseOpt otherwise.
func WhenElse(cond bool, thenOpt, elseOpt any) kit.Grouped {
	return WhenElseFn(func() bool { return cond }, thenOpt, elseOpt)
}

// WhenElseFn applies thenOpt when pred() is true and elseOpt otherwise.
func WhenElseFn(pred Predicate, thenOpt, elseOpt any) kit.Grouped {
	return kit.GroupedWhenElse(pred, []any{thenOpt}, []any{elseOpt})
}

// Unless applies setters when cond is false.
func Unless(cond bool, opts ...any) kit.Grouped {
	return When(!cond, opts...)
}

// UnlessFn applies setters when pred() is false.
func UnlessFn(pred Predicate, opts ...any) kit.Grouped {
	return WhenFn(Not(pred), opts...)
}

// OnlyIf applies the setter when check returns true. A failed check returns
// its error instead of skipping. A nil check is an error.
func OnlyIf(check Check, opt any) kit.Grouped {
	if check == nil {
		check = func() (bool, error) { return false, errNilCheck }
	}
	return kit.GroupedOnlyIf(check, []any{opt})
}

// EachKV maps each key and value to a setter and groups them.
// A nil map or factory is a no-op. Iteration order follows the map.
func EachKV[K comparable, V any](items map[K]V, f func(k K, v V) any) kit.Grouped {
	if len(items) == 0 || f == nil {
		return kit.GroupedSet(nil)
	}
	opts := make([]any, 0, len(items))
	for k, v := range items {
		opts = append(opts, f(k, v))
	}
	return Group(opts...)
}

// Each maps items to setters and groups them. A nil factory is a no-op.
func Each[T any](items []T, f func(i int, item T) any) kit.Grouped {
	if len(items) == 0 || f == nil {
		return kit.GroupedSet(nil)
	}
	opts := make([]any, 0, len(items))
	for i, item := range items {
		opts = append(opts, f(i, item))
	}
	return Group(opts...)
}

// Map maps items to setters without an index. A nil factory is a no-op.
func Map[T any](items []T, f func(item T) any) kit.Grouped {
	if f == nil {
		return kit.GroupedSet(nil)
	}
	return Each(items, func(_ int, item T) any { return f(item) })
}

// On is a [Switch] branch taken when got equals want.
func On[T comparable](got, want T, opts ...any) kit.Grouped {
	return Case(got == want, opts...)
}

// Case is a [Switch] branch taken when cond is true.
func Case(cond bool, opts ...any) kit.Grouped {
	return When(cond, opts...)
}

// CaseFn is a [Switch] branch taken when pred() is true.
func CaseFn(pred Predicate, opts ...any) kit.Grouped {
	return WhenFn(pred, opts...)
}

// Default is a [Switch] branch that always matches. Place it last.
func Default(opts ...any) kit.Grouped {
	return When(true, opts...)
}

// Switch applies the first matching [Case], [CaseFn], or [Default].
func Switch(cases ...kit.Grouped) kit.Grouped {
	return kit.GroupedSwitch(cases)
}
