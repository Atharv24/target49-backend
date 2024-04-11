package main

import (
	"fmt"

	"github.com/atharv24/target49server/tcp_server"
)

func main() {
    port := "42069"
    server := tcp_server.NewServer(":"+port)
    err := server.Start()
    if err != nil {
        fmt.Println("Error listening:", err.Error())
        return
    }
}


