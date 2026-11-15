// Package depend is the Compose depends_on condition vocabulary.
package depend

// Condition is a depends_on condition.
type Condition string

const (
	Started               Condition = "service_started"
	Healthy               Condition = "service_healthy"
	CompletedSuccessfully Condition = "service_completed_successfully"
)

// All returns every known depends_on condition.
func All() []Condition {
	return []Condition{Started, Healthy, CompletedSuccessfully}
}
