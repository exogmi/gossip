package main

import (
	"fmt"
	"log"

	"github.com/exogmi/gossip/config"
	"github.com/exogmi/gossip/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	fmt.Println("Starting Gossip IRC server...")
	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
