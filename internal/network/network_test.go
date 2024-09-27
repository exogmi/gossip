package network

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/exogmi/gossip/config"
	"github.com/exogmi/gossip/internal/state"
)

func TestNewListener(t *testing.T) {
	stateManager := &state.StateManager{}
	verbosity := config.Info

	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{"ValidAddress", "127.0.0.1:0", false},
		{"InvalidAddress", "256.0.0.1:80", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listener, err := NewListener(tt.address, stateManager, verbosity)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewListener() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && listener == nil {
				t.Errorf("NewListener() returned nil listener")
			}
			if listener != nil {
				listener.Stop()
			}
		})
	}
}

func TestListenerStartStop(t *testing.T) {
	stateManager := &state.StateManager{}
	verbosity := config.Info

	listener, err := NewListener("127.0.0.1:0", stateManager, verbosity)
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := listener.Start()
		if err != nil {
			t.Errorf("Listener.Start() error = %v", err)
		}
	}()

	// Give the listener some time to start
	time.Sleep(100 * time.Millisecond)

	listener.Stop()
	wg.Wait()
}

func TestListenerMaxConnections(t *testing.T) {
	stateManager := &state.StateManager{}
	verbosity := config.Info

	listener, err := NewListener("127.0.0.1:0", stateManager, verbosity)
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}

	maxConnections := 5
	listener.maxConnections = maxConnections

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := listener.Start()
		if err != nil {
			t.Errorf("Listener.Start() error = %v", err)
		}
	}()

	// Give the listener some time to start
	time.Sleep(100 * time.Millisecond)

	addr := listener.tcpListener.Addr().String()

	// Try to establish more than maxConnections
	for i := 0; i < maxConnections+2; i++ {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			if i >= maxConnections {
				// Expected error for connections exceeding the limit
				continue
			}
			t.Errorf("Failed to establish connection %d: %v", i+1, err)
		} else {
			defer conn.Close()
			if i >= maxConnections {
				t.Errorf("Established connection %d when it should have been rejected", i+1)
			}
		}
	}

	listener.Stop()
	wg.Wait()
}

func TestNewClientSession(t *testing.T) {
	conn, _ := net.Pipe()
	stateManager := &state.StateManager{}
	verbosity := config.Info

	session := NewClientSession(conn, stateManager, verbosity)
	if session == nil {
		t.Error("NewClientSession() returned nil")
	}
	if session.conn != conn {
		t.Error("NewClientSession() did not set the connection correctly")
	}
	if session.stateManager != stateManager {
		t.Error("NewClientSession() did not set the state manager correctly")
	}
	if session.verbosity != verbosity {
		t.Error("NewClientSession() did not set the verbosity correctly")
	}
}

func TestClientSessionReadWrite(t *testing.T) {
	client, server := net.Pipe()
	stateManager := &state.StateManager{}
	verbosity := config.Info

	session := NewClientSession(server, stateManager, verbosity)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		session.Start()
	}()

	// Write to the client
	testMessage := "TEST MESSAGE\r\n"
	_, err := client.Write([]byte(testMessage))
	if err != nil {
		t.Errorf("Failed to write to client: %v", err)
	}

	// Read from the client
	reader := bufio.NewReader(client)
	response, err := reader.ReadString('\n')
	if err != nil {
		t.Errorf("Failed to read from client: %v", err)
	}

	// Check if the response is a PING message (which is sent periodically)
	if response != "PING :server\r\n" {
		t.Errorf("Unexpected response: %s", response)
	}

	session.Stop()
	client.Close()
	wg.Wait()
}

func TestClientSessionClosure(t *testing.T) {
	client, server := net.Pipe()
	stateManager := &state.StateManager{}
	verbosity := config.Info

	session := NewClientSession(server, stateManager, verbosity)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		session.Start()
	}()

	// Close the client connection
	client.Close()

	// Wait for the session to stop
	wg.Wait()

	// Try to send a message, it should fail
	err := session.SendMessage("Test message")
	if err == nil {
		t.Error("SendMessage() should have failed after connection closure")
	}
}
