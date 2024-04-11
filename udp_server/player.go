package udp_server

import (
	"fmt"
	"net"
)

type PlayerState struct {
	ID       int
	Addr     *net.UDPAddr
	Name     string
	Position Position
	Health   int
}

type Position struct {
	x float32
	y float32
	z float32
}

func NewPlayer(id int, addr *net.UDPAddr, name string) PlayerState {
	return PlayerState{
		ID:   id,
		Addr: addr,
		Name: name,
		Position: Position{
			x: 0,
			y: 0,
			z: 0,
		},
		Health: 100,
	}
}

func (ps *PlayerState) String() string {
	return fmt.Sprintf("%d:%s", ps.ID, &ps.Position)
}
func (ps *Position) String() string {
	return fmt.Sprintf("%.3f,%.3f,%.3f", ps.x, ps.y, ps.z)
}
