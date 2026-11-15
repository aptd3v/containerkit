// Package pull is the Compose pull_policy vocabulary.
package pull

// Mode is a Compose pull_policy value.
type Mode string

const (
	Always       Mode = "always"
	Never        Mode = "never"
	Missing      Mode = "missing"
	IfNotPresent Mode = "if_not_present"
	Build        Mode = "build"
	Refresh      Mode = "refresh"
	Daily        Mode = "daily"
	Weekly       Mode = "weekly"
)

// Every returns Compose pull_policy every_<period> (for example "24h").
func Every(period string) Mode { return Mode("every_" + period) }

// All returns every fixed pull_policy (not every_<period> forms).
func All() []Mode {
	return []Mode{Always, Never, Missing, IfNotPresent, Build, Refresh, Daily, Weekly}
}
