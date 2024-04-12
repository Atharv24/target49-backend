package udp_server

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

var runMs = 5 * 1000
var serverBroadcastDelayMs = 5
var clientSendDelay = 5

var readerRunCount = runMs / serverBroadcastDelayMs
var writerRunCount = runMs / clientSendDelay

var numClients = 5

func TestStressTest(t *testing.T) {
	testServer.SetBroadcastDelay(serverBroadcastDelayMs)
	// Connect multiple clients to send data
	var clientWg sync.WaitGroup
	for i := 0; i < numClients; i++ {
		clientWg.Add(1)
		go func(id int) {
			defer clientWg.Done()
			conn, err := net.Dial("udp", fmt.Sprintf("localhost:%d", serverPort))
			if err != nil {
				t.Errorf("Failed to connect to server: %v", err)
				return
			}
			defer conn.Close()

			name := fmt.Sprintf("TEST %d", id)
			// Send a login message
			loginMessage := fmt.Sprintf("%s;%s", PLAYER_LOGIN, name)
			_, err = conn.Write([]byte(loginMessage))
			if err != nil {
				t.Errorf("Failed to send login message: %v", err)
				return
			}
			recvBuffer := make([]byte, 2048)
			conn.Read(recvBuffer)

			newPos := Position{x: float32(id), y: float32(id), z: float32(id)}
			count := 0
			for {
				// Simulate sending player state data every 5 milliseconds
				movementPacket := fmt.Sprintf("%s;%s:%d", PLAYER_STATE_MESSAGE, newPos.String(), time.Now().UnixMilli())
				_, err := conn.Write([]byte(movementPacket))
				if err != nil {
					t.Errorf("Failed to send data to server: %v", err)
					return
				}
				time.Sleep(time.Duration(clientSendDelay) * time.Millisecond)
				count++
				if count > writerRunCount {
					break
				}
			}
		}(i)
	}
	latencyInfos := [][]int64{}
	// Use a separate listener client to capture server broadcast
	var listenerWg sync.WaitGroup
	listenerWg.Add(1)
	go func() {
		defer listenerWg.Done()
		conn, err := net.Dial("udp", fmt.Sprintf("localhost:%d", serverPort))
		if err != nil {
			t.Errorf("Failed to connect to server: %v", err)
			return
		}
		defer conn.Close()

		name := "LISTENER"
		// Send a login message
		loginMessage := fmt.Sprintf("%s;%s", PLAYER_LOGIN, name)
		_, err = conn.Write([]byte(loginMessage))
		if err != nil {
			t.Errorf("Failed to send login message: %v", err)
			return
		}

		recvBuffer := make([]byte, 1024)
		count := -1
		for {
			n, err := conn.Read(recvBuffer)
			if err != nil {
				// Error handling
				fmt.Printf("error getting message from server: %s\n", err.Error())
				continue
			}

			// Process received packet
			packet := string(recvBuffer[:n])
			chunks := strings.Split(packet, ";")

			if chunks[0] != PLAYER_STATE_MESSAGE {
				continue
			}
			// Record listener client timestamp
			listenerRecvTime := time.Now().UnixMilli()
			latencyInfo := []int64{}
			for i := 1; i < len(chunks); i++ {
				chunk := chunks[i] // chunk = {ID}:{POSITION}:{TIMESTAMP}
				moreChunks := strings.Split(chunk, ":")
				if len(moreChunks) < 3 {
					t.Logf("incorrect packet structure: %s", chunk)
					continue
				}
				sentTime, err := strconv.ParseInt(moreChunks[2], 10, 64)
				if err != nil {
					// Error handling
					t.Logf("error parsing timestamp from packet: %s", chunk)
					continue
				}
				latencyInfo = append(latencyInfo, listenerRecvTime-sentTime)
			}
			latencyInfos = append(latencyInfos, latencyInfo)
			count++
			if count > readerRunCount {
				break
			}
		}
	}()

	clientWg.Wait()
	listenerWg.Wait()

	// calculate average latency for each packet and print (out of 10 clients in each packet * 100 packets)
	// also calculate 1% high, 5% high average high ping across all averages

	// Calculate average latency for each packet
	var totalLatency int64
	numPackets := int64(len(latencyInfos))
	for _, info := range latencyInfos {
		var packetLatency int64 = 0
		for _, latency := range info {
			packetLatency += latency
		}
		packetLatency /= int64(len(info))
		fmt.Printf("Average packet latency: %v\n", time.Duration(packetLatency*1000))
		totalLatency += packetLatency
	}

	avgLatency := totalLatency / numPackets
	fmt.Printf("Average Latency: %v\n", (time.Duration(avgLatency * 1000)))
}
