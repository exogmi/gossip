package network

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/exogmi/gossip/internal/models"
	"github.com/exogmi/gossip/internal/protocol"
	"github.com/exogmi/gossip/internal/state"
)

type ClientSession struct {
	conn            net.Conn
	user            *models.User
	stateManager    *state.StateManager
	protocolHandler *protocol.ProtocolHandler
	reader          *bufio.Reader
	writer          *bufio.Writer
	incoming        chan string
	outgoing        chan string
	stopChan        chan struct{}
	wg              sync.WaitGroup
}

func NewClientSession(conn net.Conn, stateManager *state.StateManager) *ClientSession {
	return &ClientSession{
		conn:            conn,
		stateManager:    stateManager,
		protocolHandler: protocol.NewProtocolHandler(stateManager),
		reader:          bufio.NewReader(conn),
		writer:          bufio.NewWriter(conn),
		incoming:        make(chan string, 100),
		outgoing:        make(chan string, 100),
		stopChan:        make(chan struct{}),
	}
}

func (cs *ClientSession) Start() {
	cs.wg.Add(3)
	go cs.readLoop()
	go cs.writeLoop()
	go cs.handleLoop()

	// Start ping-pong routine
	go cs.pingPongLoop()

	cs.wg.Wait()
}

func (cs *ClientSession) Stop() {
	close(cs.stopChan)
	cs.conn.Close()
	cs.wg.Wait()
}

func (cs *ClientSession) readLoop() {
	defer cs.wg.Done()
	for {
		select {
		case <-cs.stopChan:
			return
		default:
			line, err := cs.reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Error reading from client: %v\n", err)
				cs.Stop()
				return
			}
			cs.incoming <- line
		}
	}
}

func (cs *ClientSession) writeLoop() {
	defer cs.wg.Done()
	for {
		select {
		case <-cs.stopChan:
			return
		case msg := <-cs.outgoing:
			_, err := cs.writer.WriteString(msg + "\r\n")
			if err != nil {
				fmt.Printf("Error writing to client: %v\n", err)
				cs.Stop()
				return
			}
			cs.writer.Flush()
		}
	}
}

func (cs *ClientSession) handleLoop() {
	defer cs.wg.Done()
	for {
		select {
		case <-cs.stopChan:
			return
		case msg := <-cs.incoming:
			response, err := cs.protocolHandler.HandleMessage(cs.user, msg)
			if err != nil {
				fmt.Printf("Error handling message: %v\n", err)
				continue
			}
			if response != "" {
				cs.outgoing <- response
			}
		}
	}
}

func (cs *ClientSession) pingPongLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-cs.stopChan:
			return
		case <-ticker.C:
			cs.outgoing <- "PING :server"
			// TODO: Implement pong response handling and connection timeout
		}
	}
}

func (cs *ClientSession) SendMessage(message string) error {
	select {
	case cs.outgoing <- message:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("send message timeout")
	}
}

func (cs *ClientSession) SendNumericReply(numeric int, params ...string) error {
	message := fmt.Sprintf(":%s %03d %s", cs.conn.LocalAddr(), numeric, cs.user.Nickname)
	for _, param := range params {
		message += " " + param
	}
	return cs.SendMessage(message)
}
