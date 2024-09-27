package network

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/exogmi/gossip/config"
	"github.com/exogmi/gossip/internal/state"
)

type Listener struct {
	tcpListener  net.Listener
	stateManager *state.StateManager
	stopChan     chan struct{}
	wg           sync.WaitGroup
	verbosity    config.VerbosityLevel
}

func NewListener(address string, stateManager *state.StateManager, verbosity config.VerbosityLevel) (*Listener, error) {
	tcpListener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to create TCP listener: %w", err)
	}

	return &Listener{
		tcpListener:  tcpListener,
		stateManager: stateManager,
		stopChan:     make(chan struct{}),
		verbosity:    verbosity,
	}, nil
}

func (l *Listener) Start() error {
	defer l.wg.Wait()

	log.Printf("Listener started on %s", l.tcpListener.Addr())

	for {
		select {
		case <-l.stopChan:
			return nil
		default:
			conn, err := l.tcpListener.Accept()
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
					continue
				}
				return fmt.Errorf("error accepting connection: %w", err)
			}

			if l.verbosity >= config.Debug {
				log.Printf("New connection accepted from %s", conn.RemoteAddr())
			}

			l.wg.Add(1)
			go func() {
				defer l.wg.Done()
				session := NewClientSession(conn, l.stateManager, l.verbosity)
				session.Start()
			}()
		}
	}
}

func (l *Listener) Stop() {
	close(l.stopChan)
	l.tcpListener.Close()
	l.wg.Wait()
	log.Println("Listener stopped")
}
