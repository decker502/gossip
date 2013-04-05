package test

import (
	"fmt"
	"gossiper"
	"net/rpc/jsonrpc"
	"testing"
)

func TestConn(t *testing.T) {
	client, err := jsonrpc.Dial("tcp", "127.0.0.1:9000")
	if err != nil {
		fmt.Println("dialing:", err)
		t.Error("dialing:", err)
	}

	if client != nil {
		// Synchronous call
		args := 1
		var reply gossiper.Member
		err = client.Call("Member.Status", args, &reply)
		if err != nil {
			fmt.Println("Member status error:", err)
		} else {
			fmt.Printf("Member: %s*%d", reply.Node.Addr, reply.State)
		}
	}

}
