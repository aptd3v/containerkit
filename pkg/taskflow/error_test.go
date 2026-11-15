package taskflow

import (
	"errors"
	"testing"
)

func TestTaskError(t *testing.T) {
	var te *TaskError
	if te.Error() != "" || te.Unwrap() != nil {
		t.Fatal("nil")
	}
	if taskErr("a", nil) != nil {
		t.Fatal("nil err")
	}
	wrapped := taskErr("build", errors.New("no"))
	if taskErr("other", wrapped) != wrapped {
		t.Fatal("double wrap")
	}
	if !errors.Is(wrapped, errors.Unwrap(wrapped)) {
		t.Fatal("unwrap")
	}
	plain := &TaskError{Task: "only"}
	if plain.Error() != "task only" {
		t.Fatal(plain.Error())
	}
}
