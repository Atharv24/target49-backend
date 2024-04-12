package main

import (
	"fmt"
	"time"

	"github.com/atharv24/target49server/udp_server"
)

func main() {
	port := 42069
	broadcastDelay := 10 * time.Millisecond
	server := udp_server.NewServer(port, broadcastDelay)
	err := server.Start()
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
}
