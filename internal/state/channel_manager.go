package state

import (
	"errors"
	"sync"

	"github.com/exogmi/gossip/internal/models"
)

var (
	ErrChannelAlreadyExists = errors.New("channel already exists")
	ErrChannelNotFound      = errors.New("channel not found")
)

type ChannelManager struct {
	channels map[string]*models.Channel // Key: channel name
	mu       sync.RWMutex
}

func NewChannelManager() *ChannelManager {
	return &ChannelManager{
		channels: make(map[string]*models.Channel),
	}
}

func (cm *ChannelManager) CreateChannel(name string, creator *models.User) (*models.Channel, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.channels[name]; exists {
		return nil, ErrChannelAlreadyExists
	}

	channel := models.NewChannel(name)
	channel.AddUser(creator)
	cm.channels[name] = channel
	return channel, nil
}

func (cm *ChannelManager) RemoveChannel(name string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.channels[name]; !exists {
		return ErrChannelNotFound
	}
	delete(cm.channels, name)
	return nil
}

func (cm *ChannelManager) GetChannel(name string) (*models.Channel, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	channel, exists := cm.channels[name]
	if !exists {
		return nil, ErrChannelNotFound
	}
	return channel, nil
}

func (cm *ChannelManager) ListChannels() []*models.Channel {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	channels := make([]*models.Channel, 0, len(cm.channels))
	for _, channel := range cm.channels {
		channels = append(channels, channel)
	}
	return channels
}

func (cm *ChannelManager) JoinChannel(user *models.User, channelName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	channel, exists := cm.channels[channelName]
	if !exists {
		return ErrChannelNotFound
	}

	channel.AddUser(user)
	user.JoinChannel(channelName)
	return nil
}

func (cm *ChannelManager) LeaveChannel(user *models.User, channelName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	channel, exists := cm.channels[channelName]
	if !exists {
		return ErrChannelNotFound
	}

	channel.RemoveUser(user.Nickname)
	user.LeaveChannel(channelName)
	return nil
}

func (cm *ChannelManager) BroadcastToChannel(channel *models.Channel, message *models.Message, exclude *models.User) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for _, user := range channel.Users {
		if user != exclude {
			// TODO: Implement message delivery to user
			// This will depend on how you handle client connections
			// For now, we'll just print the message
			println("Delivering message to", user.Nickname, ":", message.Content)
		}
	}
}
