package cluster

import (
	"govclock"
)

type GossipeOverview struct {
	versions    map[Node]govclock.VClock
	unreachable []Node
}

type GossipState struct {
	version govclock.VClock
	members []Member
}
