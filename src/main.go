// gossip project main.go
package main

import (
	"cluster"
	"flag"
	"fmt"
)

func main() {
	fmt.Println("Hello World!")
	//go server.Seed()
	//server.Start()
	var seedport = flag.String("seed", "9001", "seed port")
	var selfport = flag.String("self", "9002", "self port")
	flag.Parse()
	seed := cluster.Node{"127.0.0.1", *seedport}
	self := cluster.Node{"127.0.0.1", *selfport}
	var seeds cluster.NodeSlice
	seeds = append(seeds, seed)
	cluster.Start(self, seeds)
	fmt.Println("finished!")
}
