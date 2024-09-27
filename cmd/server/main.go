package main

import (
	"log"
	"time"

	"github.com/exogmi/gossip/config"
	"github.com/exogmi/gossip/internal/server"
	"github.com/exogmi/gossip/internal/state"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set up logging based on verbosity
	switch cfg.Verbosity {
	case config.Info:
		log.SetFlags(log.Ldate | log.Ltime)
	case config.Debug:
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	case config.Trace:
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile)
	}

	// Initialize state components
	userManager := state.NewUserManager()
	messageStore := state.NewMessageStore(1000) // Store up to 1000 messages per target
	stateManager := state.NewStateManager(userManager, messageStore, "irc.gossip.local", cfg.Verbosity)

	// Start periodic cleanup of old messages
	messageStore.StartPeriodicCleanup(1 * time.Hour)

	srv, err := server.New(cfg, stateManager)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	log.Printf("Starting Gossip IRC server on %s with verbosity level %v", cfg.Address(), cfg.Verbosity)
	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
