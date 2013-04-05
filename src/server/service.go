package server

import (
	"fmt"
	"gossiper"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"time"
)

func Start() {
	ln, err := net.Listen("tcp", "127.0.0.1:9000")

	if err != nil {
		fmt.Println("network err", err)
		return
	}

	fmt.Println("server listening")
	rpc.Register(new(gossiper.Member))
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

func Seed() {

	for {
		client, err := jsonrpc.Dial("tcp", "127.0.0.1:9000")
		if err != nil {
			fmt.Println("dialing:", err)
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

		<-time.After(5 * time.Second)
	}

}
