package main

import (
	"fmt"
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

	// Initialize state components
	userManager := state.NewUserManager()
	channelManager := state.NewChannelManager()
	messageStore := state.NewMessageStore(1000) // Store up to 1000 messages per target
	stateManager := state.NewStateManager(userManager, channelManager, messageStore)

	// Start periodic cleanup of old messages
	messageStore.StartPeriodicCleanup(1 * time.Hour)

	srv, err := server.New(cfg, stateManager)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	fmt.Println("Starting Gossip IRC server...")
	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
