package cluster

type Node struct {
	Hostname string
	Port     string
}

const (
	Joining = iota
	Up
	Leaving
	Exiting
	Down
	Removed
)

type Member struct {
	Node
	State int
}

type NodeSlice []Node

func (nodes NodeSlice) Contain(value Node) int {
	for p, v := range nodes {
		if v.Hostname == value.Hostname &&
			v.Port == value.Port {
			return p
		}
	}
	return -1
}
