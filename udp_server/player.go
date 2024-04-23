package udp_server

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

type PlayerState struct {
	ID            int
	Addr          *net.UDPAddr
	Name          string
	Position      Position
	Health        int
	Score         int
	Deaths        int
	Rotation      float32
	RespawnAt     int64
	LastUpdatedAt int64
}

type Position struct {
	x float32
	y float32
	z float32
}

func RandomPosition() Position {
	return Position{
		x: rand.Float32() * 10,
		y: 10,
		z: rand.Float32() * 10,
	}
}

func NewPlayer(id int, addr *net.UDPAddr, name string) PlayerState {
	return PlayerState{
		ID:            id,
		Addr:          addr,
		Name:          name,
		Position:      RandomPosition(),
		Rotation:      0,
		Health:        MAX_HEALTH,
		Score:         0,
		Deaths:        0,
		RespawnAt:     time.Now().UnixMilli() - RESPAWN_IDLE_DELAY_MS,
		LastUpdatedAt: time.Now().UnixMilli(),
	}
}

func (ps *PlayerState) String() string {
	return fmt.Sprintf("%d:%s:%.3f:%d:%d", ps.ID, ps.Position.String(), ps.Rotation, ps.Health, ps.LastUpdatedAt)
}
func (ps Position) String() string {
	return fmt.Sprintf("%.3f,%.3f,%.3f", ps.x, ps.y, ps.z)
}
