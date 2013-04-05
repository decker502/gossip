// gossip project main.go
package main

import (
	"fmt"
	"server"
)

func main() {
	fmt.Println("Hello World!")
	//go server.Seed()
	server.Start()

}
