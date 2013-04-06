package cluster

import (
	"errors"
	"fmt"
	"govclock"
	"math/rand"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"time"
)

const (
	verNewer = iota
	verOutdated
	verSame
)
const (
	joinSuc = 1
	joinFail
)

type GossipOverview struct {
	versions    map[Node]govclock.VClock
	unreachable []Node
}

type GossipState struct {
	version govclock.VClock
	members []Member
}

type Gossiper struct {
	self           Node
	gossipState    GossipState
	isSeed         bool
	seeds          NodeSlice
	gossipOverview GossipOverview
}

type JoinResponse struct {
	JoinResult     int
	Message        string
	GossipOverview GossipOverview
	GossipState    GossipState
}
type HeartbeatRequest GossipOverview
type HeartbeatResponse struct {
	VerState    int
	GossipState GossipState
}

var vclockID = []byte("gossiper")

func (*Gossiper) Heartbeat(request HeartbeatRequest,
	response *HeartbeatResponse) error {

	response.VerState = verSame
	response.GossipState = gossiper.gossipState
	return nil
}

func (*Gossiper) JoinCluster(member *Member, response *JoinResponse) error {
	gossiper.gossipState.version = *govclock.New()
	fmt.Printf("joincluster member => %s:%s\n", member.Node.Hostname, member.Node.Port)
	member.State = Joining
	gossiper.gossipState.members = append(gossiper.gossipState.members, *member)

	response.JoinResult = joinSuc
	response.Message = "suc"
	response.GossipOverview = gossiper.gossipOverview
	response.GossipState = gossiper.gossipState

	fmt.Println("joincluster:", len(gossiper.gossipOverview.versions))
	return nil
}

func (gossiper *Gossiper) join() error {
	var addr string

	if len(gossiper.seeds) < 1 {
		return errors.New("no seeds ")
	}
	addr += gossiper.seeds[0].Hostname
	addr += ":"
	addr += gossiper.seeds[0].Port

	fmt.Printf("join addr => %s \n", addr)
	client, err := jsonrpc.Dial("tcp", addr)
	if err != nil {
		return err
	}

	if client != nil {
		// Synchronous call
		//var member Member
		//member.Node = gossiper.self
		//member.State = Joining
		member := Member{gossiper.self, Joining}
		fmt.Printf("join member => %s:%s\n", member.Node.Hostname, member.Node.Port)
		var response JoinResponse
		err = client.Call("Gossiper.JoinCluster", &member, &response)
		if err != nil {
			return err
		} else {
			fmt.Println("response: ", response.Message, len(response.GossipOverview.versions), response.JoinResult)
			gossiper.gossipOverview = response.GossipOverview
			gossiper.gossipState = response.GossipState

		}
	}
	return nil
}

func (*Gossiper) heartbeat() {
	rand.Seed(time.Now().Unix())
	for {
		members := gossiper.gossipState.members
		memberNum := len(members)
		fmt.Println("memberNum:", memberNum)
		if memberNum < 1 {
			fmt.Println("memberNum error", memberNum)
			<-time.After(5 * time.Second)
			continue
		}

		var index int = rand.Intn(memberNum)
		var addr string
		fmt.Printf("heartbeat:%d  index:%d\n", len(members), index)

		node := members[index].Node

		if node == gossiper.self {
			fmt.Printf("node:%s, self:%s\n", node.Port, gossiper.self.Port)
			<-time.After(5 * time.Second)
			continue
		}
		addr += node.Hostname
		addr += ":"
		addr += node.Port

		fmt.Printf("heartbeat addr => %s\n", addr)
		client, err := jsonrpc.Dial("tcp", addr)
		if err != nil {
			fmt.Println("dial error", err)
			<-time.After(5 * time.Second)
			continue
		}

		if client != nil {
			// Synchronous call

			var request HeartbeatRequest

			request = HeartbeatRequest(gossiper.gossipOverview)

			var response HeartbeatResponse
			err = client.Call("Gossiper.Heartbeat", request, &response)
			if err != nil {
				fmt.Println("call heartbeat error", err)
				<-time.After(5 * time.Second)
				continue
			} else {
				fmt.Println("heartbeat response: ", response.VerState, len(response.GossipState.members))

			}
		}

		<-time.After(5 * time.Second)
	}

}

func (*Gossiper) initialize() error {
	version := *govclock.New()
	var when uint64 = uint64(time.Now().Unix())
	version.Update(vclockID, when)

	gossiper.gossipState.version = version

	if gossiper.isSeed {
		gossiper.gossipOverview.versions = make(map[Node]govclock.VClock)
		gossiper.gossipOverview.versions[gossiper.self] = version

		var member Member
		member.State = Up
		member.Node = gossiper.self

		gossiper.gossipState.members = append(gossiper.gossipState.members, member)
		fmt.Println("versions len:", len(gossiper.gossipOverview.versions))

	}
	return nil
}

var gossiper Gossiper

func Start(self Node, seeds NodeSlice) {
	gossiper = *new(Gossiper)
	gossiper.self = self
	gossiper.seeds = seeds

	fmt.Println("self:", self.Hostname, self.Port)
	pos := seeds.Contain(self)
	if pos == -1 {
		gossiper.isSeed = false
		err := gossiper.join()
		if err != nil {
			fmt.Println("join fail:", err)
		}

	} else {
		fmt.Println("seed")
		gossiper.isSeed = true
	}

	err := gossiper.initialize()
	if err != nil {
		fmt.Println("gossipState init fail", err)
		return
	}

	fmt.Println("gossiper start finished")

	var addr string

	addr += self.Hostname
	addr += ":"
	addr += self.Port

	ln, err := net.Listen("tcp", addr)

	if err != nil {
		fmt.Println("network err", err)
		return
	}

	fmt.Println("server listening")

	rpc.Register(new(Gossiper))

	go gossiper.heartbeat()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("accept err", err)
			continue
		}

		fmt.Println("Accept From:", conn.RemoteAddr())

		go rpc.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}
