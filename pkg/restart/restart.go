// Package restart is the container restart-policy vocabulary.
package restart

// Policy is an Engine / Compose restart policy name.
type Policy string

const (
	No            Policy = "no"
	OnFailure     Policy = "on-failure"
	Always        Policy = "always"
	UnlessStopped Policy = "unless-stopped"
)

// All returns every known restart policy.
func All() []Policy { return []Policy{No, OnFailure, Always, UnlessStopped} }
