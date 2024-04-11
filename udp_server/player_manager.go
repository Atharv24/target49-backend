package udp_server

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

type PlayerManager struct {
	IDGenerator int
	players     sync.Map // Concurrent map for player states
	playerIDMu  sync.Mutex
}

func NewPlayerManager() *PlayerManager {
	return &PlayerManager{}
}

func (pm *PlayerManager) CreatePlayer(addr *net.UDPAddr, name string) {
	pm.playerIDMu.Lock()

	playerId := pm.IDGenerator
	pm.IDGenerator++

	pm.playerIDMu.Unlock()

	playerState := NewPlayer(playerId, addr, name)
	pm.players.Store(addr.String(), playerState)
}

func (pm *PlayerManager) UpdatePlayerState(addrStr string, newPosition Position) error {
	oldState, err := pm.GetPlayerState(addrStr)
	if err != nil {
		return err
	}
	newState := PlayerState{
		ID:       oldState.ID,
		Addr:     oldState.Addr,
		Name:     oldState.Name,
		Health:   oldState.Health,
		Position: newPosition,
	}
	// Update player's state in the concurrent map
	pm.players.Store(addrStr, newState)

	return nil
}

func (pm *PlayerManager) GetPlayerState(addrStr string) (PlayerState, error) {
	state, ok := pm.players.Load(addrStr)
	if !ok {
		return PlayerState{}, fmt.Errorf("client %s: No player state exists on server", addrStr)
	}
	playerState, ok := state.(PlayerState)
	if !ok {
		return PlayerState{}, fmt.Errorf("client %s: Unable to parse player state from map", addrStr)
	}
	return playerState, nil
}

func (pm *PlayerManager) GetAllPlayerStatesPacket() string {
	isMapEmpty := true
	strBuilder := strings.Builder{}

	strBuilder.WriteString(PLAYER_STATE_MESSAGE)

	// Iterate over the player states in the concurrent map
	pm.players.Range(func(addrStr, state interface{}) bool {
		isMapEmpty = false
		// Convert state to PlayerState type
		playerState, ok := state.(PlayerState)
		if !ok {
			// Handle type assertion error
			logger.log(LOG_LEVEL_WARNING, "Client %s: Unable to parse player state from map", addrStr)
			return true // Continue iteration
		}
		strBuilder.WriteString(fmt.Sprintf(";%s", playerState.String()))
		return true // Continue iteration
	})
	if isMapEmpty {
		return ""
	}
	return strBuilder.String()
}
