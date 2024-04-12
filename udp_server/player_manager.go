package udp_server

import (
	"fmt"
	"net"
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

func (pm *PlayerManager) CreatePlayer(addr *net.UDPAddr, name string) (PlayerState, error) {
	// check if player already logged in once
	_, ok := pm.players.Load(addr.String())
	if ok {
		return PlayerState{}, fmt.Errorf("client %s: Cant login more than once", addr.String())
	}

	pm.playerIDMu.Lock()

	pm.IDGenerator++
	playerId := pm.IDGenerator

	pm.playerIDMu.Unlock()

	playerState := NewPlayer(playerId, addr, name)

	pm.players.Store(addr.String(), playerState)

	return playerState, nil
}

func (pm *PlayerManager) UpdatePlayerState(addrStr string, newPlayerState PlayerState) error {
	oldState, err := pm.GetPlayerState(addrStr)
	if err != nil {
		return err
	}
	newState := PlayerState{
		ID:            oldState.ID,
		Addr:          oldState.Addr,
		Name:          oldState.Name,
		Health:        oldState.Health,
		Position:      newPlayerState.Position,
		LastUpdatedAt: newPlayerState.LastUpdatedAt,
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
		return PlayerState{}, fmt.Errorf("client %s: Unable to type assert player state from map", addrStr)
	}
	return playerState, nil
}

func (pm *PlayerManager) GetAllPlayerStates(skipAddr *net.UDPAddr) []PlayerState {
	states := []PlayerState{}
	// Iterate over the player states in the concurrent map
	pm.players.Range(func(addrStr, state interface{}) bool {
		if skipAddr != nil && addrStr == skipAddr.String() {
			return true
		}
		// Convert state to PlayerState type
		playerState, ok := state.(PlayerState)
		if !ok {
			// Handle type assertion error
			logger.log(LOG_LEVEL_WARNING, "Client %s: Unable to type assert player state from map", addrStr)
			return true // Continue iteration
		}
		states = append(states, playerState)
		return true // Continue iteration
	})
	// logger.log(LOG_LEVEL_DEBUG, "states read: %d", len(states))
	return states
}
