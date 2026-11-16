package types

import composeapi "github.com/docker/compose/v5/pkg/api"

// Event status values from the Compose runtime. They are constants, so the
// generated type catalog does not carry them.
const (
	Working = composeapi.Working
	Done    = composeapi.Done
	Warning = composeapi.Warning
	Error   = composeapi.Error

	StatusCreating = composeapi.StatusCreating
	StatusCreated  = composeapi.StatusCreated
)
