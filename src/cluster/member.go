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
