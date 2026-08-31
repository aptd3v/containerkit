package client

import "github.com/docker/docker/api/types/swarm"

// SwarmInit is options for SwarmInit.
type SwarmInit struct {
	ListenAddr       string
	AdvertiseAddr    string
	DataPathAddr     string
	DataPathPort     uint32
	ForceNewCluster  bool
	Spec             swarm.Spec
	AutoLockManagers bool
	Availability     swarm.NodeAvailability
	DefaultAddrPool  []string
	SubnetSize       uint32
}

func (o *SwarmInit) apply() swarm.InitRequest {
	if o == nil {
		return swarm.InitRequest{}
	}
	return swarm.InitRequest{
		ListenAddr:       o.ListenAddr,
		AdvertiseAddr:    o.AdvertiseAddr,
		DataPathAddr:     o.DataPathAddr,
		DataPathPort:     o.DataPathPort,
		ForceNewCluster:  o.ForceNewCluster,
		Spec:             o.Spec,
		AutoLockManagers: o.AutoLockManagers,
		Availability:     o.Availability,
		DefaultAddrPool:  o.DefaultAddrPool,
		SubnetSize:       o.SubnetSize,
	}
}

// SwarmJoin is options for SwarmJoin.
type SwarmJoin struct {
	ListenAddr    string
	AdvertiseAddr string
	DataPathAddr  string
	RemoteAddrs   []string
	JoinToken     string
	Availability  swarm.NodeAvailability
}

func (o *SwarmJoin) apply() swarm.JoinRequest {
	if o == nil {
		return swarm.JoinRequest{}
	}
	return swarm.JoinRequest{
		ListenAddr:    o.ListenAddr,
		AdvertiseAddr: o.AdvertiseAddr,
		DataPathAddr:  o.DataPathAddr,
		RemoteAddrs:   o.RemoteAddrs,
		JoinToken:     o.JoinToken,
		Availability:  o.Availability,
	}
}
