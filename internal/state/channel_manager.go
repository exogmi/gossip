package state

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/exogmi/gossip/internal/models"
)

var (
	ErrChannelAlreadyExists = errors.New("channel already exists")
	ErrChannelNotFound      = errors.New("channel not found")
)

type ChannelManager struct {
	channels     map[string]*models.Channel // Key: channel name
	mu           sync.RWMutex
	serverName   string
	stateManager *StateManager
}

func NewChannelManager(serverName string, stateManager *StateManager) *ChannelManager {
	return &ChannelManager{
		channels:     make(map[string]*models.Channel),
		serverName:   serverName,
		stateManager: stateManager,
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

func (cm *ChannelManager) JoinChannel(user *models.User, channelName string, key string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	channel, exists := cm.channels[channelName]
	if !exists {
		return ErrChannelNotFound
	}

	// Check if the channel has a key and if the provided key is correct
	if channel.Key != "" && channel.Key != key {
		return fmt.Errorf("cannot join channel: incorrect key")
	}

	// Check if the user is banned
	userMask := fmt.Sprintf("%s!%s@%s", user.Nickname, user.Username, user.Host)
	for _, banMask := range channel.BanList {
		if matchesMask(userMask, banMask) {
			log.Printf("User %s attempted to join channel %s but is banned", user.Nickname, channelName)
			return fmt.Errorf("cannot join channel: you're banned")
		}
	}

	wasInChannel := user.IsInChannel(channelName)
	if !wasInChannel {
		channel.AddUser(user)
		user.JoinChannel(channelName)

		// If this is the first user, make them an operator
		if len(channel.Users) == 1 {
			channel.Operators[user.Nickname] = true
		}
	}

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

	// Send updated user list to all users in the channel
	for _, u := range channel.Users {
		u.BroadcastToSessions(fmt.Sprintf(":%s 353 %s = %s :%s", cm.serverName, u.Nickname, channelName, strings.Join(userList, " ")))
		u.BroadcastToSessions(fmt.Sprintf(":%s 366 %s %s :End of /NAMES list", cm.serverName, u.Nickname, channelName))
	}

	// Replay missed messages if the user was already in the channel
	if wasInChannel {
		missedMessages, err := cm.stateManager.MessageStore.GetMessagesSince(channelName, user.LastDisconnect)
		if err != nil {
			log.Printf("Error retrieving missed messages for user %s in channel %s: %v", user.Nickname, channelName, err)
		} else {
			for _, msg := range missedMessages {
				formattedMsg := fmt.Sprintf(":%s!%s@%s PRIVMSG %s :%s", msg.Sender.Nickname, msg.Sender.Username, msg.Sender.Host, channelName, msg.Content)
				user.BroadcastToSessions(formattedMsg)
			}
		}
	}

	log.Printf("User %s joined channel %s", user.Nickname, channelName)

	return nil
}

func matchesMask(str, mask string) bool {
	// Simple wildcard matching
	// This is a basic implementation and might need to be improved for more complex IRC mask matching
	regexPattern := "^" + strings.ReplaceAll(strings.ReplaceAll(mask, ".", "\\."), "*", ".*") + "$"
	matched, _ := regexp.MatchString(regexPattern, str)
	return matched
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

	// Remove user from operators list if they were an operator
	delete(channel.Operators, user.Nickname)

	// If the channel is empty after the user leaves, remove it
	if len(channel.Users) == 0 {
		delete(cm.channels, channelName)
	}

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
