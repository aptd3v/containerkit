package client

import (
	"net/netip"

	"github.com/moby/moby/api/types/swarm"
	sdk "github.com/moby/moby/client"
)

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

func (o *SwarmInit) apply() (sdk.SwarmInitOptions, error) {
	if o == nil {
		return sdk.SwarmInitOptions{}, nil
	}
	pool := make([]netip.Prefix, 0, len(o.DefaultAddrPool))
	for _, cidr := range o.DefaultAddrPool {
		prefix, err := netip.ParsePrefix(cidr)
		if err != nil {
			return sdk.SwarmInitOptions{}, err
		}
		pool = append(pool, prefix)
	}
	return sdk.SwarmInitOptions{
		ListenAddr:       o.ListenAddr,
		AdvertiseAddr:    o.AdvertiseAddr,
		DataPathAddr:     o.DataPathAddr,
		DataPathPort:     o.DataPathPort,
		ForceNewCluster:  o.ForceNewCluster,
		Spec:             o.Spec,
		AutoLockManagers: o.AutoLockManagers,
		Availability:     o.Availability,
		DefaultAddrPool:  pool,
		SubnetSize:       o.SubnetSize,
	}, nil
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

func (o *SwarmJoin) apply() sdk.SwarmJoinOptions {
	if o == nil {
		return sdk.SwarmJoinOptions{}
	}
	return sdk.SwarmJoinOptions{
		ListenAddr:    o.ListenAddr,
		AdvertiseAddr: o.AdvertiseAddr,
		DataPathAddr:  o.DataPathAddr,
		RemoteAddrs:   o.RemoteAddrs,
		JoinToken:     o.JoinToken,
		Availability:  o.Availability,
	}
}

// SwarmUpdate is options for SwarmUpdate.
type SwarmUpdate struct {
	Version                swarm.Version
	Spec                   swarm.Spec
	RotateWorkerToken      bool
	RotateManagerToken     bool
	RotateManagerUnlockKey bool
}

func (o *SwarmUpdate) apply() sdk.SwarmUpdateOptions {
	if o == nil {
		return sdk.SwarmUpdateOptions{}
	}
	return sdk.SwarmUpdateOptions{
		Version:                o.Version,
		Spec:                   o.Spec,
		RotateWorkerToken:      o.RotateWorkerToken,
		RotateManagerToken:     o.RotateManagerToken,
		RotateManagerUnlockKey: o.RotateManagerUnlockKey,
	}
}

// SwarmUnlock is options for SwarmUnlock.
type SwarmUnlock struct {
	Key string
}

func (o *SwarmUnlock) apply() sdk.SwarmUnlockOptions {
	if o == nil {
		return sdk.SwarmUnlockOptions{}
	}
	return sdk.SwarmUnlockOptions{Key: o.Key}
}
