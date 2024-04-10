package server

import (
	"fmt"
	"net"
	"sync"
)

type clientManager struct {
	clientCounter int
	clients       map[int]net.Conn
	clientsMu     sync.RWMutex
}

func NewClientManager() *clientManager {
	return &clientManager{
		clientCounter: 0,
		clients:       make(map[int]net.Conn),
	}
}

func (cm *clientManager) GetClient(clientID int) net.Conn {
	return cm.clients[clientID]
}

func (cm *clientManager) sendMessage(clientId int, message string) error {
	client := cm.clients[clientId]
	_, err := client.Write([]byte(message))

	return err
}

func (cm *clientManager) broadcastMessage (fromClient int, message string) {
	for toClient := range cm.clients {
		if toClient == fromClient {
			continue
		}
		err := cm.sendMessage(toClient, message)
		if err != nil {
			fmt.Printf("Error sending message to client %d: %s\n", toClient, err.Error())
			continue
		}
	}
}

func (cm *clientManager) BroadcastClientPos(clientID int, pos []float32) {
	cm.clientsMu.Lock()
	defer cm.clientsMu.Unlock()
	// broadcast
	message, err := GetPositionString(clientID, pos)
	if err != nil {
		fmt.Printf("Error creating position string for Client %d: %s\n", clientID, err.Error())
		return
	}
	cm.broadcastMessage(clientID, message)
}

func (cm *clientManager) RegisterClient(conn net.Conn) (int, error) {
	cm.clientsMu.Lock()
	defer cm.clientsMu.Unlock()

	cm.clientCounter++
	newClientID := cm.clientCounter
	cm.clients[newClientID] = conn
	
	fmt.Println("Registered new client:", newClientID)
	
	cm.initializeExistingClientsForNewClient(newClientID)
	cm.broadcastNewClient(newClientID)
	
	return newClientID, nil
}

func (cm *clientManager) UnregisterClient(clientID int) {
	cm.clientsMu.Lock()
	defer cm.clientsMu.Unlock()

	client := cm.clients[clientID]
	client.Close()

	delete(cm.clients, clientID)

	// notify other clients
  cm.broadcastClientRelease(clientID)	
	fmt.Printf("Removed Client %v\n", clientID)
}

func (cm *clientManager) initializeExistingClientsForNewClient(newClientID int) {
	existingClients := []string{}
	for existingClient := range cm.clients {
		if existingClient == newClientID {
			continue
		}
		existingClients = append(existingClients, fmt.Sprint(existingClient))
	}
	if len(existingClients) == 0 {
		return
	}
	message := GetInitializeString(existingClients)
	err := cm.sendMessage(newClientID, message)
	if err != nil {
		fmt.Printf("Error initializing existing clients for %d: %s\n", newClientID, err.Error())
	}
}

func (cm *clientManager) broadcastNewClient(clientID int) {
	message := GetInitializeString([]string{fmt.Sprint(clientID)})
	cm.broadcastMessage(clientID, message)
}

func (cm *clientManager) broadcastClientRelease(clientID int) {
	message := GetReleaseString(clientID)
	cm.broadcastMessage(clientID, message)
}
