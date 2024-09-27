package protocol

import (
	"fmt"

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

func (ph *ProtocolHandler) HandleCommand(user *models.User, command *Command) (string, error) {
	switch command.Name {
	case "NICK":
		return ph.handleNickCommand(user, command.Params)
	case "USER":
		return ph.handleUserCommand(user, command.Params)
	case "JOIN":
		return ph.handleJoinCommand(user, command.Params)
	case "PRIVMSG":
		return ph.handlePrivmsgCommand(user, command.Params)
	default:
		return "", fmt.Errorf("unknown command: %s", command.Name)
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
