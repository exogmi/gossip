package state

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/exogmi/gossip/internal/models"
)

var (
	ErrChannelAlreadyExists = errors.New("channel already exists")
	ErrChannelNotFound      = errors.New("channel not found")
)

type ChannelManager struct {
	channels   map[string]*models.Channel // Key: channel name
	mu         sync.RWMutex
	serverName string
}

func NewChannelManager(serverName string) *ChannelManager {
	return &ChannelManager{
		channels:   make(map[string]*models.Channel),
		serverName: serverName,
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

	// Broadcast JOIN message to all users in the channel
	joinMsg := fmt.Sprintf(":%s!%s@%s JOIN %s", user.Nickname, user.Username, user.Host, channelName)
	for _, u := range channel.Users {
		u.BroadcastToSessions(joinMsg)
	}

	// Send channel topic to the joining user
	user.BroadcastToSessions(fmt.Sprintf(":%s 332 %s %s :%s", cm.serverName, user.Nickname, channelName, channel.Topic))
	
	// Send user list to the joining user
	userList := channel.GetUserList()
	user.BroadcastToSessions(fmt.Sprintf(":%s 353 %s = %s :%s", cm.serverName, user.Nickname, channelName, strings.Join(userList, " ")))
	user.BroadcastToSessions(fmt.Sprintf(":%s 366 %s %s :End of /NAMES list", cm.serverName, user.Nickname, channelName))

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
			var formattedMsg string
			switch message.Type {
			case models.ChannelMessage:
				formattedMsg = fmt.Sprintf(":%s!%s@%s PRIVMSG %s :%s", message.Sender.Nickname, message.Sender.Username, message.Sender.Host, channel.Name, message.Content)
			case models.ServerMessage:
				formattedMsg = message.Content
			}
			user.BroadcastToSessions(formattedMsg)
		}
	}
}
