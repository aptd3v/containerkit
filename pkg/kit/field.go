package kit

// field distinguishes omitted values from explicitly supplied zeros.
type field[T any] struct {
	set bool
	val T
}

func (f *field[T]) assign(v T) {
	f.set = true
	f.val = v
}

func (f field[T]) ok() bool { return f.set }
