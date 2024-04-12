package main

import (
	"fmt"

	"github.com/atharv24/target49server/udp_server"
)

func main() {
	port := 42069
	broadcastDelayMs := 5
	server := udp_server.NewServer(port, broadcastDelayMs)
	err := server.Start()
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
}
