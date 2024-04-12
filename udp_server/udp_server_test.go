package udp_server

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testServer *server
var serverPort = 42069

func TestMain(m *testing.M) {
	// Setup server once before running tests
	broadcastDelayMs := 5000 // large broadcast delay initially
	testServer = NewServer(serverPort, broadcastDelayMs)
	go func() {
		if err := testServer.Start(); err != nil {
			panic(err)
		}
	}()
	defer func() {
		testServer.quitCh <- struct{}{}
	}()

	// Give some time for the server to start
	time.Sleep(100 * time.Millisecond)

	// Run the tests
	result := m.Run()

	// Cleanup code, if any

	// Return the test result
	os.Exit(result)
}

func TestHandlePlayerLogin(t *testing.T) {
	// Connect to the server
	conn, err := net.Dial("udp", "localhost:42069")
	if err != nil {
		t.Errorf("Failed to connect to server: %v", err)
		return
	}
	defer conn.Close()

	name := "Atharv"
	// Send a login message
	loginMessage := fmt.Sprintf("%s;%s", PLAYER_LOGIN, name)
	_, err = conn.Write([]byte(loginMessage))
	if err != nil {
		t.Errorf("Failed to send login message: %v", err)
		return
	}

	// Receive response
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		t.Errorf("Failed to read response: %v", err)
		return
	}
	// Check response
	actualResponse := string(buffer[:n])
	chunks := strings.Split(actualResponse, ";")
	assert.GreaterOrEqual(t, len(chunks), 2)
	assert.Equal(t, INITIAL_MESSAGE, chunks[0])

	moreChunks := strings.Split(chunks[1], ":")
	assert.Equal(t, len(moreChunks), 3)

	pos := Position{}
	assert.Equal(t, pos.String(), moreChunks[1])
}
func TestHandleTwoPlayerLogin(t *testing.T) {
	// Connect to the server
	conn, err := net.Dial("udp", "localhost:42069")
	if err != nil {
		t.Errorf("Failed to connect to server: %v", err)
		return
	}
	defer conn.Close()

	name := "Atthu"
	// Send a login message
	loginMessage := fmt.Sprintf("%s;%s", PLAYER_LOGIN, name)
	_, err = conn.Write([]byte(loginMessage))
	if err != nil {
		t.Errorf("Failed to send login message: %v", err)
		return
	}

	// Receive response
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		t.Errorf("Failed to read response: %v", err)
		return
	}

	actualResponse := string(buffer[:n])
	chunks := strings.Split(actualResponse, ";")
	assert.GreaterOrEqual(t, len(chunks), 2)
	assert.Equal(t, INITIAL_MESSAGE, chunks[0])

	moreChunks := strings.Split(chunks[1], ":")
	assert.Equal(t, 3, len(moreChunks))

	id1, err := strconv.Atoi(moreChunks[0])
	assert.Nil(t, err)

	pos := Position{}
	assert.Equal(t, pos.String(), moreChunks[1])

	conn2, err := net.Dial("udp", "localhost:42069")
	if err != nil {
		t.Errorf("Failed to connect to server: %v", err)
		return
	}
	defer conn2.Close()

	name2 := "Ath"
	// Send a login message
	loginMessage2 := fmt.Sprintf("%s;%s", PLAYER_LOGIN, name2)
	_, err = conn2.Write([]byte(loginMessage2))
	if err != nil {
		t.Errorf("Failed to send login message: %v", err)
		return
	}

	// Receive response
	n2, err := conn2.Read(buffer)
	if err != nil {
		t.Errorf("Failed to read response: %v", err)
		return
	}

	actualResponse2 := string(buffer[:n2])

	chunks2 := strings.Split(actualResponse2, ";")
	assert.GreaterOrEqual(t, len(chunks2), 3)
	assert.Equal(t, INITIAL_MESSAGE, chunks2[0])
	moreChunks2 := strings.Split(chunks2[1], ":")

	assert.Equal(t, 3, len(moreChunks2))

	id2, err := strconv.Atoi(moreChunks2[0])
	assert.Equal(t, nil, err)
	assert.NotEqual(t, id1, id2)

	assert.Equal(t, pos.String(), moreChunks2[1])

	id1Found := false
	for i := 2; i < len(chunks2); i++ {
		moreChunks3 := strings.Split(chunks2[i], ":")
		assert.Equal(t, 3, len(moreChunks3))
		assert.Equal(t, pos.String(), moreChunks3[1])
		id3, err := strconv.Atoi(moreChunks3[0])
		assert.Nil(t, err)
		if id3 == id1 {
			id1Found = true
			break
		}
	}
	assert.Equal(t, true, id1Found)

	n, err = conn.Read(buffer)
	if err != nil {
		t.Errorf("Failed to read response: %v", err)
		return
	}

	actualResponse = string(buffer[:n])

	chunks = strings.Split(actualResponse, ";")
	assert.Equal(t, 2, len(chunks))
	assert.Equal(t, NEW_PLAYER, chunks[0])

	moreChunks = strings.Split(chunks[1], ":")
	assert.Equal(t, 3, len(moreChunks))

	id, err := strconv.Atoi(moreChunks[0])
	assert.Nil(t, err)

	assert.Equal(t, id2, id)
	assert.Equal(t, pos.String(), moreChunks[1])
}

func TestBroadcasting(t *testing.T) {
	// add one more player to enable broadcasting
	conn2, err := net.Dial("udp", "localhost:42069")
	if err != nil {
		t.Errorf("Failed to connect to server: %v", err)
		return
	}
	defer conn2.Close()

	// Send a login message
	name := "Atthu"
	loginMessage := fmt.Sprintf("%s;%s", PLAYER_LOGIN, name)
	_, err = conn2.Write([]byte(loginMessage))
	if err != nil {
		t.Errorf("Failed to send login message: %v", err)
		return
	}

	conn, err := net.Dial("udp", "localhost:42069")
	if err != nil {
		t.Errorf("Failed to connect to server: %v", err)
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte(loginMessage))
	if err != nil {
		t.Errorf("Failed to send login message: %v", err)
		return
	}

	// Receive response
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		t.Errorf("Failed to read response: %v", err)
		return
	}

	initPacket := string(buffer[:n])
	selfState := strings.Split(initPacket, ";")[1]
	id := selfState[0]

	testServer.SetBroadcastDelay(10)
	// read 10 state packets
	for i := 0; i < 10; i++ {
		n, err = conn.Read(buffer)
		if err != nil {
			t.Errorf("Failed to read response: %v", err)
			return
		}
		packet := string(buffer[:n])
		chunks := strings.Split(packet, ";")
		assert.Equal(t, PLAYER_STATE_MESSAGE, chunks[0])

		idFound := false
		for _, chunk := range chunks {
			if chunk[0] == id {
				idFound = true
				break
			}
		}
		assert.True(t, idFound)
	}
	newPos := Position{
		1,
		1,
		1,
	}
	movementPacket := fmt.Sprintf("%s;%s:%.3f", PLAYER_STATE_MESSAGE, newPos.String(), float64(time.Now().UnixMilli()/1000))
	conn.Write([]byte(movementPacket))

	// read 10 state packets
	for i := 0; i < 10; i++ {
		n, err = conn.Read(buffer)
		if err != nil {
			t.Errorf("Failed to read response: %v", err)
			return
		}
		packet := string(buffer[:n])
		chunks := strings.Split(packet, ";")
		assert.Equal(t, PLAYER_STATE_MESSAGE, chunks[0])
		idFound := false
		for i := 1; i < len(chunks); i++ {
			moreChunks := strings.Split(chunks[i], ":")
			if moreChunks[0][0] == id {
				idFound = true
				assert.Equal(t, newPos.String(), moreChunks[1])
				break
			}
		}
		assert.True(t, idFound)
	}
}
