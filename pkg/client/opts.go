package client

import (
	"github.com/docker/docker/api/types/filters"
)

// Filter is a key/value filter passed to list and prune operations.
type Filter struct {
	Key   string
	Value string
}

func argsFromFilters(fs []Filter) filters.Args {
	a := filters.NewArgs()
	for _, f := range fs {
		a.Add(f.Key, f.Value)
	}
	return a
}

// WaitCondition is a container state to wait for.
type WaitCondition string

const (
	WaitConditionNotRunning WaitCondition = "not-running"
	WaitConditionNextExit   WaitCondition = "next-exit"
	WaitConditionRemoved    WaitCondition = "removed"
)
