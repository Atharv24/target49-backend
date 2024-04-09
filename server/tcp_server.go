package server

import (
	"fmt"
	"net"
	"time"
)

type server struct {
	listenAddr    string
	ln            net.Listener
	quitCh        chan struct{}
	parser        parser
	clientManager clientManager
}

func NewServer(addr string) *server {
	return &server{
		listenAddr:    addr,
		parser:        *NewParser(),
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

	// start the parser
	go s.parser.Start()

	// start accepting connections
	go s.acceptLoop()
	go s.processLoop()
	go s.writeLoop()

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
			fmt.Printf("Error registering client:%s", err.Error())
			continue
		}

		// start reading data from conn
		go s.readLoop(clientID)
	}
}

func (s *server) readLoop(clientID int) {
	defer s.clientManager.UnregisterClient(clientID)
	client := s.clientManager.GetClient(clientID)
	buf := make([]byte, 64)
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
		s.parser.messageCh <- data
	}
}

func (s *server) writeLoop() {
	ticker := time.NewTicker(100 * time.Millisecond) // Create a ticker with a duration of 0.1 seconds
	defer ticker.Stop()                              // Stop the ticker when the function exits

	for {
		select {
		case <-ticker.C:
			s.clientManager.SyncClientPos() // Call SyncClientPos() every time the ticker ticks
		case <-s.quitCh:
			return // Exit the function if quitCh is closed
		}
	}
}

func (s *server) processLoop() {
	for cmd := range s.parser.cmdCh {
		if cmd.cmd == POSITION {
			pos, err := ParseCoorString(cmd.data)	
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			s.clientManager.UpdateClientPos(cmd.clientId, pos)
		}
	}
}

