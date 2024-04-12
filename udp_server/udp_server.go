package udp_server

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type server struct {
	conn            *net.UDPConn
	port            int
	quitCh          chan struct{}
	quitReceive     chan struct{}
	quitBroadcast   chan struct{}
	broadcastTicker *time.Ticker
	broadcastLock   sync.Mutex
	playerManager   *PlayerManager
}

func NewServer(port int, broadcastDelayMs int) *server {
	return &server{
		port:            port,
		playerManager:   NewPlayerManager(),
		broadcastTicker: time.NewTicker(time.Duration(broadcastDelayMs) * time.Millisecond),
		quitCh:          make(chan struct{}),
		quitReceive:     make(chan struct{}),
		quitBroadcast:   make(chan struct{}),
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
	go s.broadcastPlayerStates()

	// Wait for server to be stopped
	<-s.quitCh

	// Signal both goroutines to stop
	close(s.quitReceive)
	close(s.quitBroadcast)
	s.broadcastTicker.Stop()

	return nil
}

func (s *server) receiveMessages() {
	for {
		select {
		case <-s.quitReceive:
			return
		default:
			buf := make([]byte, 1024)
			n, addr, err := s.conn.ReadFromUDP(buf)
			if err != nil {
				logger.log(LOG_LEVEL_WARNING, "Error reading from UDP:%s", err.Error())
				continue
			}

			// start processing message
			go s.processMessage(addr, buf[:n])
		}
	}
}

func (s *server) broadcastPlayerStates() {
	for {
		select {
		case <-s.quitBroadcast:
			return
		case <-s.broadcastTicker.C:
			// logger.log(LOG_LEVEL_DEBUG, "BROADCAST TICK")
			if playerStates := s.playerManager.GetAllPlayerStates(nil); len(playerStates) > 1 {
				// go s.calculateBroadcastDelay(playerStates)

				broadcastPacket := parser.EncodePlayerStatesForBroadcast(playerStates)
				playerAddrs := []*net.UDPAddr{}
				for _, ps := range playerStates {
					playerAddrs = append(playerAddrs, ps.Addr)
				}
				s.broadcastPacket(playerAddrs, broadcastPacket)
			}
		}
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
		s.handlePlayerStateUpdate(addr, msg.data)
	case PLAYER_LOGIN:
		s.handlePlayerLogin(addr, msg.data)
	default:
		logger.log(LOG_LEVEL_WARNING, "Unknown message type: %s", data)
	}
}

func (s *server) handlePlayerStateUpdate(addr *net.UDPAddr, data string) {
	ps, err := parser.ParsePlayerState(data)
	if err != nil {
		logger.log(LOG_LEVEL_WARNING, "Unable to parse player state from packet (%s): %s", data, err)
		return
	}
	addrStr := addr.String()
	err = s.playerManager.UpdatePlayerState(addrStr, ps)
	if err != nil {
		logger.log(LOG_LEVEL_WARNING, err.Error())
	}
}

func (s *server) sendPacket(addr *net.UDPAddr, packet string) {
	s.conn.WriteToUDP([]byte(packet), addr)
}

func (s *server) broadcastPacket(addrs []*net.UDPAddr, packet string) {
	for _, addr := range addrs {
		s.sendPacket(addr, packet)
	}
}

func (s *server) handlePlayerLogin(addr *net.UDPAddr, data string) {
	name := parser.ParseLoginMessage(data)
	newPlayerState, err := s.playerManager.CreatePlayer(addr, name)
	if err != nil {
		logger.log(LOG_LEVEL_WARNING, err.Error())
		return
	}

	// Send all logged in players
	existingPlayerStates := s.playerManager.GetAllPlayerStates(newPlayerState.Addr)
	initPacket := parser.EncodePlayerStatesForInit(newPlayerState, existingPlayerStates)

	// logger.log(LOG_LEVEL_DEBUG, "Player %d: Init packet (%s)", newPlayerState.ID, initPacket)
	s.sendPacket(newPlayerState.Addr, initPacket)

	existingPlayerAddrs := []*net.UDPAddr{}
	for _, ps := range existingPlayerStates {
		existingPlayerAddrs = append(existingPlayerAddrs, ps.Addr)
	}
	// broadcast to all players that new player is here
	packet := parser.EncodePlayerStateForInit(newPlayerState)

	// logger.log(LOG_LEVEL_DEBUG, "Player %d: Broadcast packet (%s)", newPlayerState.ID, packet)
	s.broadcastPacket(existingPlayerAddrs, packet)

	logger.log(LOG_LEVEL_INFO, "Player %d logged in: %s", newPlayerState.ID, newPlayerState.Name)
}

func (s *server) SetBroadcastDelay(newDelayMs int) {
	// Lock to prevent race conditions while updating broadcastDelay
	s.broadcastLock.Lock()
	defer s.broadcastLock.Unlock()

	s.quitBroadcast <- struct{}{}

	// Stop the current ticker
	s.broadcastTicker.Stop()

	// Start a new ticker with the updated delay
	s.broadcastTicker = time.NewTicker(time.Duration(newDelayMs) * time.Millisecond)
	go s.broadcastPlayerStates()
}

// func (s *server) calculateBroadcastDelay(playerStates []PlayerState) {
// 	current_ms := time.Now().UnixMilli()
// 	for _, ps := range playerStates {
// 		latency := current_ms - ps.LastUpdatedAt

// 		if latency > 100 {
// 			logger.log(LOG_LEVEL_DEBUG, "[%v] Player %d: High latency %dms", time.Now(), ps.ID, latency)
// 		}
// 	}
// }
