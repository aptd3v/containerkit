package client

import sdk "github.com/moby/moby/client"

// Filter is a key/value filter passed to list and prune operations.
type Filter struct {
	Key   string
	Value string
}

func argsFromFilters(fs []Filter) sdk.Filters {
	out := make(sdk.Filters)
	for _, f := range fs {
		out.Add(f.Key, f.Value)
	}
	return out
}

// WaitCondition is a container state to wait for.
type WaitCondition string

const (
	WaitConditionNotRunning WaitCondition = "not-running"
	WaitConditionNextExit   WaitCondition = "next-exit"
	WaitConditionRemoved    WaitCondition = "removed"
)
