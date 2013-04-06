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
}
type HeartbeatRequest GossipOverview
type HeartbeatResponse struct {
	verState    int
	gossipState GossipState
}

var vclockID = []byte("gossiper")

func (*GossipOverview) Heartbeat(request HeartbeatRequest,
	response *HeartbeatResponse) error {

	response.verState = verSame
	response.gossipState = gossiper.gossipState
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

	client, err := jsonrpc.Dial("tcp", addr)
	if err != nil {
		return err
	}

	if client != nil {
		// Synchronous call
		var member Member
		member.Node = gossiper.self
		member.State = Joining

		var response JoinResponse
		err = client.Call("Gossiper.JoinCluster", member, &response)
		if err != nil {
			return err
		} else {
			fmt.Println("response: ", response.Message, len(response.GossipOverview.versions), response.JoinResult)
			gossiper.gossipOverview = response.GossipOverview

		}
	}
	return nil
}

func (*Gossiper) JoinCluster(member Member, response *JoinResponse) error {
	gossiper.gossipState.version = *govclock.New()

	member.State = Joining
	gossiper.gossipState.members = append(gossiper.gossipState.members, member)

	response.JoinResult = joinSuc
	response.Message = "suc"
	response.GossipOverview = gossiper.gossipOverview
	fmt.Println("joincluster:", len(gossiper.gossipOverview.versions))
	return nil
}

func (*Gossiper) initialize() error {
	version := *govclock.New()
	var when uint64 = uint64(time.Now().Unix())
	version.Update(vclockID, when)

	gossiper.gossipState.version = version

	if gossiper.isSeed {
		gossiper.gossipOverview.versions = make(map[Node]govclock.VClock)
		gossiper.gossipOverview.versions[gossiper.self] = version
		fmt.Println("versions len:", len(gossiper.gossipOverview.versions))

	}
	return nil
}

var gossiper Gossiper

func (*Gossiper) heartbeat() {
	rand.Seed(time.Now().Unix())
	for {
		versionLen := len(gossiper.gossipOverview.versions)
		
		if versionLen < 1 {
			fmt.Println("version len error", versionLen)
		}
		var index int = rand.intn(versionLen - 1)
		
		fmt.Println("heartbeat:", len(gossiper.gossipOverview.versions))
		<-time.After(5 * time.Second)
	}

}

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
