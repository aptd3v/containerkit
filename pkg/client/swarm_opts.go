package client

import sdk "github.com/moby/moby/client"

// SwarmInit is [sdk.SwarmInitOptions].
type SwarmInit sdk.SwarmInitOptions

func (o *SwarmInit) apply() sdk.SwarmInitOptions { return sdk.SwarmInitOptions(apply(o)) }

// SwarmJoin is [sdk.SwarmJoinOptions].
type SwarmJoin sdk.SwarmJoinOptions

func (o *SwarmJoin) apply() sdk.SwarmJoinOptions { return sdk.SwarmJoinOptions(apply(o)) }

// SwarmUpdate is [sdk.SwarmUpdateOptions].
type SwarmUpdate sdk.SwarmUpdateOptions

func (o *SwarmUpdate) apply() sdk.SwarmUpdateOptions { return sdk.SwarmUpdateOptions(apply(o)) }

// SwarmUnlock is [sdk.SwarmUnlockOptions].
type SwarmUnlock sdk.SwarmUnlockOptions

func (o *SwarmUnlock) apply() sdk.SwarmUnlockOptions { return sdk.SwarmUnlockOptions(apply(o)) }
