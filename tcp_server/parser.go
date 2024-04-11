package tcp_server

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type message struct {
	clientID int
	cmd      string
	data     string
}

const (
	POSITION		= "P"
	INITIALIZE  = "N"
	RELEASE 		= "D"
)

func ParseMessage(clientID int, data []byte) (message, error) {
	if len(data) < 1 {
		return message{}, errors.New("empty message")
	}
	parsedData := string(data)
	chunks := strings.Split(parsedData, ":")

	m := message{
		clientID: clientID,
		cmd:      chunks[0],
	}
	if len(chunks) > 1 {
		m.data = chunks[1]
	}
	return m, nil
}

func GetPositionString(clientID int, coors []float32) (string, error) {
	if len(coors) != 3 {
		return "", fmt.Errorf("coors length should be 3")
	}
	return fmt.Sprintf("%s:%d:%.4f,%.4f,%.4f", POSITION, clientID, coors[0], coors[1], coors[2]), nil
}

func ParsePositionString(data string) ([]float32, error) {
	posStrs := strings.Split(data, ",")
	pos := []float32{}

	for _, posStr := range posStrs {
		posFloat, err := strconv.ParseFloat(posStr, 32)
		if err != nil {
			return []float32{}, err
		}
		pos = append(pos, float32(posFloat))
	}
	return pos, nil
}

func GetInitializeString(clientIDs []string) string{
	return fmt.Sprintf("%s:%s", INITIALIZE, strings.Join(clientIDs, ","))
}

func GetReleaseString(clientID int) string{
	return fmt.Sprintf("%s:%d", RELEASE, clientID)
}
