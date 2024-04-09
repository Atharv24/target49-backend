package server

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

type clientManager struct {
	clientCounter int
	clients       map[int]net.Conn
	clientPos     map[int][]float32
	clientsMu     sync.RWMutex
}

func NewClientManager() *clientManager {
	return &clientManager{
		clientCounter: 0,
		clients:       make(map[int]net.Conn),
		clientPos:     make(map[int][]float32),
	}
}

func (cm *clientManager) GetClient(clientID int) net.Conn {
	return cm.clients[clientID]
}

// to each client, send the positions of rest of the clients
// in the format P:{client1ID}:x,y,z:{client2ID}:x,y,z
func (cm *clientManager) SyncClientPos() {
	cm.clientsMu.Lock()
	defer cm.clientsMu.Unlock()
	if len(cm.clients) == 0 {
		return
	}
	var builder strings.Builder
	builder.WriteString("P")

	// Build positions string
	for clientID, clientPos := range cm.clientPos {
		coorStr, err := GetCoorString(clientPos)
		if err != nil {
			fmt.Printf("Error converting Coordinates to String:%s", err.Error())
			return
		}
		builder.WriteString(fmt.Sprintf(":%d#%s", clientID, coorStr))
	}
	positionsStr := builder.String()

	// Send positions to each client
	for clientID := range cm.clients {
		if err := cm.SendMessage(clientID, positionsStr); err != nil {
			fmt.Printf("Error syncing client positions to client:%d\n", clientID)
		}
	}
}

func (cm *clientManager) SendMessage(clientId int, message string) error {
	client := cm.clients[clientId]
	_, err := client.Write([]byte(message))

	if err != nil {
		return err
	}

	return nil
}

func (cm *clientManager) UpdateClientPos(clientID int, pos []float32) {
	cm.clientsMu.Lock()
	defer cm.clientsMu.Unlock()
	cm.clientPos[clientID] = pos
}

func (cm *clientManager) RegisterClient(conn net.Conn) (int, error) {
	cm.clientsMu.Lock()
	defer cm.clientsMu.Unlock()

	clientID := cm.clientCounter
	cm.clientCounter++

	cm.clients[clientID] = conn

	clientPos := []float32{0.0, 1.0, 0.0}
	cm.clientPos[clientID] = clientPos

	coorStr, err := GetCoorString(clientPos)
	if err != nil {
		return 0, fmt.Errorf("error converting Coordinates to String:%s", err.Error())
	}
	// Initial client setup
	message := fmt.Sprintf("I:%v#%s", clientID, coorStr)

	err = cm.SendMessage(clientID, message)
	if err != nil {
		return 0, fmt.Errorf("error sending Client ID to client:%s", err.Error())
	}
	fmt.Println("Registered new client:", clientID)

	cm.BroadcastNewClient(clientID)

	return clientID, nil
}

func (cm *clientManager) UnregisterClient(clientID int) {
	cm.clientsMu.Lock()

	client := cm.clients[clientID]
	client.Close()

	delete(cm.clients, clientID)
	delete(cm.clientPos, clientID)
	cm.clientsMu.Unlock()

	fmt.Printf("Removed Client %v\n", clientID)
}

func (cm *clientManager) InitializeClientsForNewClient(clientID int) error {
	
	for _clientID, clientPos := range cm.clientPos {
		coorStr, err := GetCoorString(clientPos)
		if err != nil {
			return fmt.Errorf("error getting coors string:%s", err.Error())
		}
		message := fmt.Sprintf("N:%d#%s", clientID, coorStr)
		if _clientID == clientID {
			continue
		}
		err := cm.SendMessage(_clientID, message)
		if err != nil {
			fmt.Printf("Error sending message to client %d: %s", _clientID, err.Error())
			continue
		}
	}
	return nil
}

func (cm *clientManager) BroadcastNewClient(clientID int) error {
	clientPos := cm.clientPos[clientID]
	coorStr, err := GetCoorString(clientPos)
	if err != nil {
		return fmt.Errorf("error getting coors string:%s", err.Error())
	}
	message := fmt.Sprintf("N:%d#%s", clientID, coorStr)
	for _clientID := range cm.clients {
		if _clientID == clientID {
			continue
		}
		err := cm.SendMessage(_clientID, message)
		if err != nil {
			fmt.Printf("Error sending message to client %d: %s", _clientID, err.Error())
			continue
		}
	}
	return nil
}
