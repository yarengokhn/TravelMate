package tests

import (
	"bufio"
	"net"
	"testing"
	"time"
	"travel-platform/internal/chat"

	"github.com/stretchr/testify/assert"
)

func TestTCPServer_Connectivity(t *testing.T) {
	// Start server in a goroutine
	address := "127.0.0.1:9091" // Use different port for testing
	server := chat.NewServer(address)

	go func() {
		_ = server.Start()
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Connect as a client
	conn, err := net.Dial("tcp", address)
	assert.NoError(t, err)
	defer conn.Close()

	// Read greeting
	reader := bufio.NewReader(conn)
	greeting, err := reader.ReadString('\n')
	assert.NoError(t, err)
	assert.Contains(t, greeting, "Welcome to TravelMate Chat")
}
