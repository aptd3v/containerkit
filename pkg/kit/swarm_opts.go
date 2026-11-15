package kit

// ListenAddr sets the swarm listen address (for example 0.0.0.0:2377).
func ListenAddr(addr string) ListenAddrOpt { return ListenAddrOpt{addr: addr} }

// ListenAddrOpt is the value returned by [ListenAddr].
// It implements [EngineSwarmInitOption] and [EngineSwarmJoinOption].
type ListenAddrOpt struct{ addr string }

func (o ListenAddrOpt) applyEngineSwarmInitOption(a *engineSwarmInitAct) error {
	a.opts.ListenAddr = o.addr
	return nil
}

func (o ListenAddrOpt) applyEngineSwarmJoinOption(a *engineSwarmJoinAct) error {
	a.opts.ListenAddr = o.addr
	return nil
}

// AdvertiseAddr sets the address swarm advertises to other nodes
// (for example 10.0.0.11:2377).
func AdvertiseAddr(addr string) AdvertiseAddrOpt { return AdvertiseAddrOpt{addr: addr} }

// AdvertiseAddrOpt is the value returned by [AdvertiseAddr].
// It implements [EngineSwarmInitOption] and [EngineSwarmJoinOption].
type AdvertiseAddrOpt struct{ addr string }

func (o AdvertiseAddrOpt) applyEngineSwarmInitOption(a *engineSwarmInitAct) error {
	a.opts.AdvertiseAddr = o.addr
	return nil
}

func (o AdvertiseAddrOpt) applyEngineSwarmJoinOption(a *engineSwarmJoinAct) error {
	a.opts.AdvertiseAddr = o.addr
	return nil
}

// DataPathAddr sets the address used for data-path traffic.
func DataPathAddr(addr string) DataPathAddrOpt { return DataPathAddrOpt{addr: addr} }

// DataPathAddrOpt is the value returned by [DataPathAddr].
// It implements [EngineSwarmInitOption] and [EngineSwarmJoinOption].
type DataPathAddrOpt struct{ addr string }

func (o DataPathAddrOpt) applyEngineSwarmInitOption(a *engineSwarmInitAct) error {
	a.opts.DataPathAddr = o.addr
	return nil
}

func (o DataPathAddrOpt) applyEngineSwarmJoinOption(a *engineSwarmJoinAct) error {
	a.opts.DataPathAddr = o.addr
	return nil
}

// ForceNewCluster forces a new cluster from the current state on [InitSwarm].
func ForceNewCluster(v bool) ForceNewClusterOpt { return ForceNewClusterOpt{v: v} }

// ForceNewClusterOpt is the value returned by [ForceNewCluster].
// It implements [EngineSwarmInitOption].
type ForceNewClusterOpt struct{ v bool }

func (o ForceNewClusterOpt) applyEngineSwarmInitOption(a *engineSwarmInitAct) error {
	a.opts.ForceNewCluster = o.v
	return nil
}

// JoinToken sets the swarm join token for [JoinSwarm].
func JoinToken(token string) JoinTokenOpt { return JoinTokenOpt{token: token} }

// JoinTokenOpt is the value returned by [JoinToken].
// It implements [EngineSwarmJoinOption].
type JoinTokenOpt struct{ token string }

func (o JoinTokenOpt) applyEngineSwarmJoinOption(a *engineSwarmJoinAct) error {
	a.opts.JoinToken = o.token
	return nil
}

// RemoteAddrs adds manager addresses to contact when joining a swarm
// (for example 10.0.0.11:2377).
func RemoteAddrs(addrs ...string) RemoteAddrsOpt { return RemoteAddrsOpt{addrs: addrs} }

// RemoteAddrsOpt is the value returned by [RemoteAddrs].
// It implements [EngineSwarmJoinOption].
type RemoteAddrsOpt struct{ addrs []string }

func (o RemoteAddrsOpt) applyEngineSwarmJoinOption(a *engineSwarmJoinAct) error {
	a.opts.RemoteAddrs = append(a.opts.RemoteAddrs, o.addrs...)
	return nil
}

var (
	_ EngineSwarmInitOption = ListenAddr("0.0.0.0:2377")
	_ EngineSwarmInitOption = AdvertiseAddr("10.0.0.11:2377")
	_ EngineSwarmInitOption = DataPathAddr("10.0.0.11")
	_ EngineSwarmInitOption = ForceNewCluster(true)
	_ EngineSwarmJoinOption = ListenAddr("0.0.0.0:2377")
	_ EngineSwarmJoinOption = AdvertiseAddr("10.0.0.12:2377")
	_ EngineSwarmJoinOption = DataPathAddr("10.0.0.12")
	_ EngineSwarmJoinOption = JoinToken("token")
	_ EngineSwarmJoinOption = RemoteAddrs("10.0.0.11:2377")
)
