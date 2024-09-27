package network

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/exogmi/gossip/config"
	"github.com/exogmi/gossip/internal/models"
	"github.com/exogmi/gossip/internal/protocol"
	"github.com/exogmi/gossip/internal/state"
	"github.com/google/uuid"
)

type ClientSession struct {
	conn            net.Conn
	user            *models.User
	stateManager    *state.StateManager
	protocolParser  *protocol.ProtocolParser
	protocolHandler *protocol.ProtocolHandler
	reader          *bufio.Reader
	writer          *bufio.Writer
	incoming        chan string
	outgoing        chan string
	stopChan        chan struct{}
	wg              sync.WaitGroup
	verbosity       config.VerbosityLevel
	clientID        string
}

// Ensure ClientSession implements the models.ClientSession interface
var _ models.ClientSession = (*ClientSession)(nil)

func (cs *ClientSession) SetUser(user *models.User) {
	cs.user = user
	user.ClientSession = cs
}

func NewClientSession(conn net.Conn, stateManager *state.StateManager, verbosity config.VerbosityLevel) *ClientSession {
	return &ClientSession{
		conn:            conn,
		stateManager:    stateManager,
		protocolParser:  protocol.NewProtocolParser(),
		protocolHandler: protocol.NewProtocolHandler(stateManager),
		reader:          bufio.NewReader(conn),
		writer:          bufio.NewWriter(conn),
		incoming:        make(chan string, 100),
		outgoing:        make(chan string, 100),
		stopChan:        make(chan struct{}),
		verbosity:       verbosity,
		clientID:        uuid.New().String(),
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
	if cs.verbosity >= config.Debug {
		log.Printf("Client session stopped for %s", cs.clientID)
	}
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
				log.Printf("Error reading from client %s: %v", cs.clientID, err)
				cs.Stop()
				return
			}
			if cs.verbosity >= config.Trace {
				log.Printf("Received from client %s: %s", cs.clientID, line)
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
				log.Printf("Error writing to client %s: %v", cs.clientID, err)
				cs.Stop()
				return
			}
			cs.writer.Flush()
			if cs.verbosity >= config.Trace {
				log.Printf("Sent to client %s: %s", cs.clientID, msg)
			}
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
			command, err := cs.protocolParser.Parse(msg)
			if err != nil {
				log.Printf("Error parsing message from client %s: %v", cs.clientID, err)
				continue
			}
			if cs.protocolHandler == nil {
				log.Printf("Error: ProtocolHandler is nil for client %s", cs.clientID)
				continue
			}
			if cs.verbosity >= config.Debug {
				log.Printf("Handling command for client %s: %s", cs.clientID, command.Name)
			}
			response, err := cs.protocolHandler.HandleCommand(cs.user, command)
			if err != nil {
				log.Printf("Error handling command for client %s: %v", cs.clientID, err)
				continue
			}
			if response != "" {
				cs.outgoing <- response
			}
			if cs.user == nil && (command.Name == "NICK" || command.Name == "USER") {
				user := cs.protocolHandler.GetUser()
				if user != nil {
					cs.SetUser(user)
				}
			}
			if command.Name == "QUIT" {
				cs.Stop()
				return
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
			if cs.verbosity >= config.Trace {
				log.Printf("Sent PING to client %s", cs.clientID)
			}
			// TODO: Implement pong response handling and connection timeout
		}
	}
}

func (cs *ClientSession) SendMessage(message string) error {
	select {
	case cs.outgoing <- message:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("send message timeout for client %s", cs.clientID)
	}
}

func (cs *ClientSession) SendNumericReply(numeric int, params ...string) error {
	message := fmt.Sprintf(":%s %03d %s", cs.conn.LocalAddr(), numeric, cs.user.Nickname)
	for _, param := range params {
		message += " " + param
	}
	return cs.SendMessage(message)
}
