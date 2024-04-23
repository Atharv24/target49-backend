package udp_server

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type PlayerManager struct {
	IDGenerator     int
	players         sync.Map // Concurrent map for player states
	playerIDMu      sync.RWMutex
	playerIDAddrMap map[int]string
}

func NewPlayerManager() *PlayerManager {
	return &PlayerManager{
		playerIDAddrMap: make(map[int]string),
	}
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
	pm.playerIDAddrMap[playerId] = addr.String()

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
	if oldState.LastUpdatedAt > newPlayerState.LastUpdatedAt {
		return fmt.Errorf("stale player state for Player %d", oldState.ID)
	}
	if newPlayerState.LastUpdatedAt-oldState.RespawnAt < RESPAWN_IDLE_DELAY_MS {
		return fmt.Errorf("player state updated before respawn delay")
	}

	newState := PlayerState{
		ID:            oldState.ID,
		Addr:          oldState.Addr,
		Name:          oldState.Name,
		Health:        oldState.Health,
		Rotation:      newPlayerState.Rotation,
		Position:      newPlayerState.Position,
		Score:         oldState.Score,
		Deaths:        oldState.Deaths,
		RespawnAt:     oldState.RespawnAt,
		LastUpdatedAt: newPlayerState.LastUpdatedAt,
	}
	// Update player's state in the concurrent map
	pm.players.Store(addrStr, newState)

	return nil
}

func (pm *PlayerManager) HandlePlayerShot(receiverID int, shooterAddr *net.UDPAddr, lastUpdatedAt int64) *net.UDPAddr {
	pm.playerIDMu.RLock()
	receieverAddr, ok := pm.playerIDAddrMap[receiverID]
	if !ok {
		logger.warn("Player %d doesnt exist", receiverID)
		return nil
	}

	pm.playerIDMu.RUnlock()
	receieverState, err := pm.GetPlayerState(receieverAddr)
	if err != nil {
		logger.warn("Unable to get Player %d state: %s", receiverID, err.Error())
	}
	if receieverState.Health > 1 {
		newState := PlayerState{
			ID:            receieverState.ID,
			Addr:          receieverState.Addr,
			Name:          receieverState.Name,
			Health:        receieverState.Health - 1,
			Rotation:      receieverState.Rotation,
			Score:         receieverState.Score,
			Deaths:        receieverState.Deaths,
			Position:      receieverState.Position,
			RespawnAt:     receieverState.RespawnAt,
			LastUpdatedAt: lastUpdatedAt,
		}
		pm.players.Store(receieverAddr, newState)
		return nil
	} else {
		// Handle player respawn
		logger.info("Player %d died, respawning", receiverID)
		newState := PlayerState{
			ID:            receieverState.ID,
			Addr:          receieverState.Addr,
			Name:          receieverState.Name,
			Health:        MAX_HEALTH,
			Rotation:      0.0,
			Score:         receieverState.Score,
			Deaths:        receieverState.Deaths + 1,
			Position:      RandomPosition(),
			RespawnAt:     time.Now().UnixMilli(),
			LastUpdatedAt: lastUpdatedAt,
		}
		pm.players.Store(receieverAddr, newState)

		shooterState, err := pm.GetPlayerState(shooterAddr.String())
		if err != nil {
			logger.warn("Unable to get Client %s state: %s", shooterAddr.String(), err.Error())
		}
		newShooterState := PlayerState{
			ID:            shooterState.ID,
			Addr:          shooterState.Addr,
			Name:          shooterState.Name,
			Health:        shooterState.Health,
			Rotation:      shooterState.Rotation,
			Score:         shooterState.Score + 1,
			Deaths:        shooterState.Deaths,
			Position:      shooterState.Position,
			RespawnAt:     shooterState.RespawnAt,
			LastUpdatedAt: shooterState.LastUpdatedAt,
		}
		pm.players.Store(shooterAddr.String(), newShooterState)
		return receieverState.Addr
	}
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
			logger.warn("Client %s: Unable to type assert player state from map", addrStr)
			return true // Continue iteration
		}
		states = append(states, playerState)
		return true // Continue iteration
	})
	// logger.log(LOG_LEVEL_DEBUG, "states read: %d", len(states))
	return states
}
