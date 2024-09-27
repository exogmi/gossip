package protocol

import (
	"fmt"
	"strings"

	"github.com/exogmi/gossip/internal/models"
	"github.com/exogmi/gossip/internal/state"
)

type ProtocolHandler struct {
	stateManager *state.StateManager
}

func NewProtocolHandler(stateManager *state.StateManager) *ProtocolHandler {
	return &ProtocolHandler{
		stateManager: stateManager,
	}
}

func (ph *ProtocolHandler) HandleMessage(user *models.User, message string) (string, error) {
	parts := strings.SplitN(strings.TrimSpace(message), " ", 2)
	command := strings.ToUpper(parts[0])

	switch command {
	case "NICK":
		return ph.handleNickCommand(user, parts[1:])
	case "USER":
		return ph.handleUserCommand(user, parts[1:])
	case "JOIN":
		return ph.handleJoinCommand(user, parts[1:])
	case "PRIVMSG":
		return ph.handlePrivmsgCommand(user, parts[1:])
	default:
		return "", fmt.Errorf("unknown command: %s", command)
	}
}

func (ph *ProtocolHandler) handleNickCommand(user *models.User, params []string) (string, error) {
	// TODO: Implement NICK command handling
	return "", nil
}

func (ph *ProtocolHandler) handleUserCommand(user *models.User, params []string) (string, error) {
	// TODO: Implement USER command handling
	return "", nil
}

func (ph *ProtocolHandler) handleJoinCommand(user *models.User, params []string) (string, error) {
	// TODO: Implement JOIN command handling
	return "", nil
}

func (ph *ProtocolHandler) handlePrivmsgCommand(user *models.User, params []string) (string, error) {
	// TODO: Implement PRIVMSG command handling
	return "", nil
}
