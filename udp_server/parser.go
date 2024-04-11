package udp_server

import (
	"fmt"
	"strings"
)

type Parser struct{}

type message struct {
	messageType string
	data        string
}

const (
	PLAYER_STATE_MESSAGE = "S"
	PLAYER_LOGIN         = "L"
)

var parser Parser // Package-level variable to hold the logger instance

func init() {
	parser = Parser{}
}

func (p *Parser) ParseMessage(data []byte) (message, error) {
	if len(data) < 1 {
		return message{}, fmt.Errorf("empty packet")
	}
	parsedData := string(data)
	chunks := strings.Split(parsedData, ";")
	if len(chunks) < 2 {
		return message{}, fmt.Errorf("missing type or data in packet")
	}
	message := message{
		messageType: chunks[0],
		data:        chunks[1],
	}
	return message, nil
}

func (p *Parser) ParsePlayerState(newStateStr string) (Position, error) {
	// newStateStr = "0.000,0.000,0.000"
	if len(newStateStr) < 5*3+2 {
		return Position{}, fmt.Errorf("invalid length of state packet")
	}

	var pos Position
	fmt.Sscanf(newStateStr, "%f,%f,%f", &pos.x, &pos.y, &pos.z)

	return pos, nil
}

func (p *Parser) ParseLoginMessage(loginData string) string {
	return loginData
}
