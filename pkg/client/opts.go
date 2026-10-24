package client

import (
	"github.com/moby/moby/api/types/container"
	sdk "github.com/moby/moby/client"
)

// Filters is [sdk.Filters].
type Filters = sdk.Filters

// WaitCondition is [container.WaitCondition].
type WaitCondition = container.WaitCondition

const (
	WaitConditionNotRunning = container.WaitConditionNotRunning
	WaitConditionNextExit   = container.WaitConditionNextExit
	WaitConditionRemoved    = container.WaitConditionRemoved
)

// ConsoleSize is [sdk.ConsoleSize].
type ConsoleSize = sdk.ConsoleSize

func apply[T any](o *T) T {
	if o == nil {
		var z T
		return z
	}
	return *o
}
