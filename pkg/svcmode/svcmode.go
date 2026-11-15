// Package svcmode is the Compose deploy.mode / Swarm service-mode vocabulary.
package svcmode

// Mode is deploy.mode (replicated, global, replicated-job, global-job).
type Mode string

const (
	Replicated    Mode = "replicated"
	Global        Mode = "global"
	ReplicatedJob Mode = "replicated-job"
	GlobalJob     Mode = "global-job"
)

// All returns every known service deploy mode.
func All() []Mode { return []Mode{Replicated, Global, ReplicatedJob, GlobalJob} }
