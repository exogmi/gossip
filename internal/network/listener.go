package network

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"

	"github.com/exogmi/gossip/config"
	"github.com/exogmi/gossip/internal/state"
)

type Listener struct {
	tcpListener    net.Listener
	sslListener    net.Listener
	stateManager   *state.StateManager
	stopChan       chan struct{}
	wg             sync.WaitGroup
	verbosity      config.VerbosityLevel
	maxConnections int
	ActiveConns    int32
	useSSL         bool
}

func NewListener(address string, sslAddress string, stateManager *state.StateManager, verbosity config.VerbosityLevel, useSSL bool, sslCertFile, sslKeyFile string) (*Listener, error) {
	l := &Listener{
		stateManager:   stateManager,
		stopChan:       make(chan struct{}),
		verbosity:      verbosity,
		maxConnections: 1000,
		ActiveConns:    0,
		useSSL:         useSSL,
	}

	var err error
	l.tcpListener, err = net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to create TCP listener: %w", err)
	}

	if useSSL {
		cert, err := tls.LoadX509KeyPair(sslCertFile, sslKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load SSL certificates: %w", err)
		}

		config := &tls.Config{Certificates: []tls.Certificate{cert}}
		l.sslListener, err = tls.Listen("tcp", sslAddress, config)
		if err != nil {
			return nil, fmt.Errorf("failed to create SSL listener: %w", err)
		}
	}

	return l, nil
}

func (l *Listener) Start() error {
	defer l.wg.Wait()

	log.Printf("Listener started on %s", l.tcpListener.Addr())
	if l.useSSL {
		log.Printf("SSL Listener started on %s", l.sslListener.Addr())
	}

	l.wg.Add(1)
	go l.acceptLoop(l.tcpListener)

	if l.useSSL {
		l.wg.Add(1)
		go l.acceptLoop(l.sslListener)
	}

	<-l.stopChan
	return nil
}

func (l *Listener) acceptLoop(listener net.Listener) {
	defer l.wg.Done()

	for {
		select {
		case <-l.stopChan:
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
					continue
				}
				log.Printf("Error accepting connection: %v", err)
				return
			}

			if atomic.LoadInt32(&l.ActiveConns) >= int32(l.maxConnections) {
				log.Printf("Maximum connections reached, rejecting connection from %s", conn.RemoteAddr())
				conn.Close()
				continue
			}

			atomic.AddInt32(&l.ActiveConns, 1)

			if l.verbosity >= config.Debug {
				log.Printf("New connection accepted from %s", conn.RemoteAddr())
			}

			l.wg.Add(1)
			go func() {
				defer l.wg.Done()
				defer atomic.AddInt32(&l.ActiveConns, -1)
				session := NewClientSession(conn, l.stateManager, l.verbosity)
				session.Start()
			}()
		}
	}
}

func (l *Listener) Stop() {
	close(l.stopChan)
	l.tcpListener.Close()
	if l.useSSL {
		l.sslListener.Close()
	}
	l.wg.Wait()
	log.Println("Listener stopped")
}
