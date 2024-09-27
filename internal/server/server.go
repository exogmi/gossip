package server

import (
	"fmt"

	"github.com/exogmi/gossip/config"
	"github.com/exogmi/gossip/internal/network"
	"github.com/exogmi/gossip/internal/state"
)

// Server represents the IRC server
type Server struct {
	config       *config.Config
	stateManager *state.StateManager
	listener     *network.Listener
}

// New creates a new Server instance
func New(cfg *config.Config, stateManager *state.StateManager) (*Server, error) {
	listener, err := network.NewListener(cfg.Address(), stateManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	return &Server{
		config:       cfg,
		stateManager: stateManager,
		listener:     listener,
	}, nil
}

// Run starts the server
func (s *Server) Run() error {
	fmt.Printf("Server listening on %s\n", s.config.Address())
	return s.listener.Start()
}

// Stop gracefully stops the server
func (s *Server) Stop() {
	s.listener.Stop()
}
