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
	// S;{POS} from client, S;{ID}:{POS};{ID}:{POS} from server
	PLAYER_STATE_MESSAGE = "S"

	// L;{NAME} from client
	PLAYER_LOGIN = "L"

	// I;{NEW_PLAYER_ID}:{NEW_POS};{ID1}:{POS1};{ID2}:{POS2} from server
	INITIAL_MESSAGE = "I"

	// N;{NEW_PLAYER_ID}:{NEW_POS} from server
	NEW_PLAYER = "N"
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

func (p *Parser) EncodePlayerStatesForBroadcast(playerStates []PlayerState) string {
	if len(playerStates) < 2 {
		return ""
	}
	strBuilder := strings.Builder{}

	strBuilder.WriteString(PLAYER_STATE_MESSAGE)
	for _, ps := range playerStates {
		strBuilder.WriteString(fmt.Sprintf(";%s", ps.String()))
	}
	return strBuilder.String()
}

func (p *Parser) EncodePlayerStatesForInit(
	newPlayerState PlayerState,
	existingPlayersState []PlayerState,
) string {
	strBuilder := strings.Builder{}

	strBuilder.WriteString(fmt.Sprintf("%s;%s", INITIAL_MESSAGE, newPlayerState.String()))
	for _, ps := range existingPlayersState {
		strBuilder.WriteString(fmt.Sprintf(";%s", ps.String()))
	}
	return strBuilder.String()
}

func (p *Parser) EncodePlayerStateForInit(
	newPlayerState PlayerState,
) string {
	return fmt.Sprintf("%s;%s", NEW_PLAYER, newPlayerState.String())
}
