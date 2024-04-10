package server

import (
	"fmt"
	"net"
)

type server struct {
	listenAddr    string
	ln            net.Listener
	quitCh        chan struct{}
	clientManager clientManager
}

func NewServer(addr string) *server {
	return &server{
		listenAddr:    addr,
		clientManager: *NewClientManager(),
	}
}

func (s *server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()
	fmt.Println("Server listening on port 42069")
	s.ln = ln

	// start accepting connections
	go s.acceptLoop()

	<-s.quitCh

	return nil
}

func (s *server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			fmt.Println("Error accepting conn:", err.Error())
			continue
		}
		// register client
		clientID, err := s.clientManager.RegisterClient(conn)
		if err != nil {
			fmt.Printf("Error registering client:%s\n", err.Error())
			continue
		}

		// start reading data from conn
		go s.readLoop(clientID, conn)
	}
}

func (s *server) readLoop(clientID int, client net.Conn) {
	defer s.clientManager.UnregisterClient(clientID)
	buf := make([]byte, 1024)
	for {
		n, err := client.Read(buf)
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Printf("Client %v disconnected, removing client\n", clientID)
				return
			}
			fmt.Println("Error reading:", err.Error())
			continue
		}
		data := buf[:n]
		go s.processMessage(clientID, data)
	}
}

func (s *server) processMessage(clientID int, data []byte) {
	cmd, err := ParseMessage(clientID, data)
	if err != nil {
		fmt.Printf("Error parsing message: %s\n", err.Error())
		return
	}
	switch cmd.cmd {
	case POSITION:
		pos, err := ParsePositionString(cmd.data)
		if err != nil {
			fmt.Printf("error parsing coordinate strings: %s\n", err.Error())
			return
		}
		s.clientManager.BroadcastClientPos(cmd.clientID, pos)
	default:
		fmt.Printf("Received unknown packet from Client %d: %s", clientID, string(data))
	}
}
