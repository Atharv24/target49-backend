package server

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type parser struct {
	messageCh chan []byte // Channel to receive messages from
	cmdCh     chan message
}

type message struct {
	clientId int
	cmd      string
	data     string
}

const (
	POSITION string = "P"
)

func NewParser() *parser {
	return &parser{
		messageCh: make(chan []byte, 10),
		cmdCh:     make(chan message, 10),
	}
}

func (p *parser) Start() {
	defer close(p.cmdCh)
	defer close(p.messageCh)

	for data := range p.messageCh {
		// Parse and process the incoming message
		message, err := p.parseMessage(data)
		if err != nil {
			fmt.Println("Error parsing message:", err.Error())
			continue
		}

		p.cmdCh <- message
	}
}

func (p *parser) parseMessage(data []byte) (message, error) {
	if len(data) < 1 {
		return message{}, errors.New("empty message")
	}
	parsedData := string(data)
	chunks := strings.Split(parsedData, ":")

	clientId, err := strconv.Atoi(chunks[0])
	if err != nil {
		errMsg := fmt.Sprintf("error parsing ClientID from message: %s", err.Error())
		return message{}, errors.New(errMsg)
	}
	m := message{
		clientId: clientId,
		cmd:      chunks[1],
	}
	if len(chunks) > 2 {
		m.data = chunks[2]
	}
	return m, nil
}

func GetCoorString(coors []float32) (string, error) {
	if len(coors) != 3 {
		return "", fmt.Errorf("coors length should be 3")
	}
	return fmt.Sprintf("%.2f,%.2f,%.2f", coors[0], coors[1], coors[2]), nil
}

func ParseCoorString(data string) ([]float32, error) {
	posStrs := strings.Split(data, ",")
	pos := []float32{}

	for _, posStr := range posStrs {
		posFloat, err := strconv.ParseFloat(posStr, 32)
		if err != nil {
			return []float32{}, fmt.Errorf("error parsing POS for client:%s", err.Error())
		}
		pos = append(pos, float32(posFloat))
	}
	return pos, nil
}
