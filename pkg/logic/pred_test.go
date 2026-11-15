package logic_test

import (
	"testing"

	"github.com/aptd3v/containerkit/pkg/logic"
)

func TestAndOr(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		pred logic.Predicate
		want bool
	}{
		{name: "and all true", pred: logic.And(true, true), want: true},
		{name: "and one false", pred: logic.And(true, false), want: false},
		{name: "and empty", pred: logic.And(), want: true},
		{name: "or any true", pred: logic.Or(false, true), want: true},
		{name: "or all false", pred: logic.Or(false, false), want: false},
		{name: "or empty", pred: logic.Or(), want: false},
		{
			name: "andfn skips nil",
			pred: logic.AndFn(nil, func() bool { return true }),
			want: true,
		},
		{
			name: "orfn skips nil",
			pred: logic.OrFn(nil, func() bool { return false }, func() bool { return true }),
			want: true,
		},
		{
			name: "composed with whenfn",
			pred: logic.OrFn(logic.And(false, true), logic.Or(true)),
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.pred(); got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}

func TestAndOrShortCircuit(t *testing.T) {
	t.Parallel()
	called := false
	if logic.AndFn(func() bool { return false }, func() bool { called = true; return true })() {
		t.Fatal("andfn")
	}
	if called {
		t.Fatal("andfn called later predicate")
	}
	called = false
	if !logic.OrFn(func() bool { return true }, func() bool { called = true; return false })() {
		t.Fatal("orfn")
	}
	if called {
		t.Fatal("orfn called later predicate")
	}
}
