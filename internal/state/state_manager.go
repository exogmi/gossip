package state

import (
	"github.com/exogmi/gossip/internal/models"
)

// StateManager serves as the central point for accessing all state-related operations
type StateManager struct {
	UserManager    *UserManager
	ChannelManager *ChannelManager
	MessageStore   *MessageStore
	ServerName     string
}

// NewStateManager creates a new StateManager instance
func NewStateManager(userManager *UserManager, channelManager *ChannelManager, messageStore *MessageStore, serverName string) *StateManager {
	return &StateManager{
		UserManager:    userManager,
		ChannelManager: channelManager,
		MessageStore:   messageStore,
		ServerName:     serverName,
	}
}

// GetUser retrieves a user by nickname
func (sm *StateManager) GetUser(nickname string) (*models.User, error) {
	return sm.UserManager.GetUser(nickname)
}

// GetChannel retrieves a channel by name
func (sm *StateManager) GetChannel(name string) (*models.Channel, error) {
	return sm.ChannelManager.GetChannel(name)
}

// CreateUser creates a new user
func (sm *StateManager) CreateUser(nickname, username, realname, host string) (*models.User, error) {
	user := models.NewUser(nickname, username, realname, host)
	err := sm.UserManager.AddUser(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// CreateChannel creates a new channel
func (sm *StateManager) CreateChannel(name string, creator *models.User) (*models.Channel, error) {
	return sm.ChannelManager.CreateChannel(name, creator)
}

// StoreMessage stores a message
func (sm *StateManager) StoreMessage(message *models.Message) error {
	return sm.MessageStore.StoreMessage(message)
}

// GetMessages retrieves messages for a target
func (sm *StateManager) GetMessages(target string, limit int) ([]*models.Message, error) {
	return sm.MessageStore.GetMessages(target, limit)
}
