// Package role is the Swarm node-role vocabulary.
package role

// Role is a swarm node role.
type Role string

const (
	Manager Role = "manager"
	Worker  Role = "worker"
)

// All returns every known node role.
func All() []Role { return []Role{Manager, Worker} }
