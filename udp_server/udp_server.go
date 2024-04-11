package udp_server

import (
	"fmt"
	"net"
	"time"
)

type server struct {
	conn          *net.UDPConn
	port          int
	quitCh        chan struct{}
	playerManager *PlayerManager
}

func NewServer(port int) *server {
	return &server{
		port:          port,
		playerManager: NewPlayerManager(),
	}

}

func (s *server) Start() error {
	logger.setLogLevel(LOG_LEVEL_DEBUG)

	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("error resolving server address: %s", err.Error())
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("error starting server: %s", err.Error())
	}
	defer conn.Close()

	fmt.Println("Server listening on port 42069")
	s.conn = conn
	go s.receiveMessages()
	go s.broadcastMessages()

	<-s.quitCh

	return nil
}

func (s *server) receiveMessages() {
	for {
		buf := make([]byte, 1024)
		n, addr, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			logger.log(LOG_LEVEL_WARNING, "Error reading from UDP:%s", err.Error())
			continue
		}

		// start processing message
		go s.processMessage(addr, buf[:n-1])
	}
}

func (s *server) processMessage(addr *net.UDPAddr, data []byte) {
	msg, err := parser.ParseMessage(data)
	if err != nil {
		logger.log(LOG_LEVEL_WARNING, "Unable to parse packet (%s): %s", data, err)
		return
	}
	switch msg.messageType {
	case PLAYER_STATE_MESSAGE:
		pos, err := parser.ParsePlayerState(msg.data)
		if err != nil {
			logger.log(LOG_LEVEL_WARNING, "Unable to parse player state from packet (%s): %s", data, err)
			return
		}
		addrStr := addr.String()
		err = s.playerManager.UpdatePlayerState(addrStr, pos)
		if err != nil {
			logger.log(LOG_LEVEL_WARNING, err.Error())
			return
		}
	case PLAYER_LOGIN:
		name := parser.ParseLoginMessage(msg.data)
		s.playerManager.CreatePlayer(addr, name)
	default:
		logger.log(LOG_LEVEL_WARNING, "Unknown message type: %s", data)
	}
}

func (s *server) broadcastMessages() {
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		if broadcastData := s.playerManager.GetAllPlayerStatesPacket(); broadcastData != "" {
			logger.log(LOG_LEVEL_DEBUG, broadcastData)
		}
	}
}
