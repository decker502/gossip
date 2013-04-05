package gossiper

import (
	"fmt"
)

type Node struct {
	Addr string
}

const (
	Joning = iota
	Up
	Leaving
	Exiting
	Down
	Removed
)

type Member struct {
	Node  Node
	State int
}

func (*Member) Status(flag int, member *Member) error {
	fmt.Println("flag = ", flag)
	member.Node.Addr = "127.0.0.1:9001"
	member.State = Joning
	return nil
}
