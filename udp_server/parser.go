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

func (p *Parser) ParsePlayerState(newStateStr string) (PlayerState, error) {
	// newStateStr = "0.000,0.000,0.000:0.000:123123441"
	var ps PlayerState
	_, err := fmt.Sscanf(newStateStr, "%f,%f,%f:%f:%d", &ps.Position.x, &ps.Position.y, &ps.Position.z, &ps.Rotation, &ps.LastUpdatedAt)

	return ps, err
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

func (p *Parser) EncodePlayerResetMessage() string {
	return fmt.Sprintf("%s;%s:%.3f:%d", PLAYER_RESET, RandomPosition().String(), 0.0, MAX_HEALTH)
}
