package server

import (
	"fmt"
	"net"

	"github.com/exogmi/gossip/config"
)

// Server represents the IRC server
type Server struct {
	config *config.Config
}

// New creates a new Server instance
func New(cfg *config.Config) (*Server, error) {
	return &Server{
		config: cfg,
	}, nil
}

// Run starts the server
func (s *Server) Run() error {
	listener, err := net.Listen("tcp", s.config.Address())
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}
	defer listener.Close()

	fmt.Printf("Server listening on %s\n", s.config.Address())

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Printf("New connection from %s\n", conn.RemoteAddr())
	// TODO: Implement actual connection handling
}
