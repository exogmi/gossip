package models

import (
	"fmt"
	"time"
)

// Channel represents an IRC channel
type Channel struct {
	Name        string
	Topic       string
	CreatedAt   time.Time
	Users       map[string]*User
	Modes       ChannelModes
	UserLimits  int
	BanList     []string
	InviteList  []string
	Key         string
	Operators   map[string]bool
}

// ChannelModes represents the modes a channel can have
type ChannelModes struct {
	InviteOnly               bool
	Moderated                bool
	NoExternal               bool
	Private                  bool
	Secret                   bool
	TopicSettableOnlyByOps   bool
}

// NewChannel creates a new Channel instance
func NewChannel(name string) *Channel {
	return &Channel{
		Name:        name,
		Topic:       fmt.Sprintf("Welcome to %s!", name),
		CreatedAt:   time.Now(),
		Users:       make(map[string]*User),
		Modes:       ChannelModes{},
		UserLimits:  0,
		BanList:     make([]string, 0),
		InviteList:  make([]string, 0),
		Operators:   make(map[string]bool),
	}
}

// AddUser adds a user to the channel
func (c *Channel) AddUser(user *User) {
	c.Users[user.Nickname] = user
}

// RemoveUser removes a user from the channel
func (c *Channel) RemoveUser(nickname string) {
	delete(c.Users, nickname)
}

// SetTopic sets the channel topic
func (c *Channel) SetTopic(topic string) {
	c.Topic = topic
}

// SetMode sets a mode for the channel
func (c *Channel) SetMode(mode string, value bool) error {
	switch mode {
	case "i": // invite-only
		c.Modes.InviteOnly = value
	case "m": // moderated
		c.Modes.Moderated = value
	case "n": // no external messages
		c.Modes.NoExternal = value
	case "p": // private
		c.Modes.Private = value
	case "s": // secret
		c.Modes.Secret = value
	case "t": // topic settable only by ops
		c.Modes.TopicSettableOnlyByOps = value
	default:
		return fmt.Errorf("unknown mode: %s", mode)
	}
	return nil
}

// IsBanned checks if a user mask is banned from the channel
func (c *Channel) IsBanned(userMask string) bool {
	for _, ban := range c.BanList {
		if ban == userMask {
			return true
		}
	}
	return false
}

// IsInvited checks if a nickname is invited to the channel
func (c *Channel) IsInvited(nickname string) bool {
	for _, invite := range c.InviteList {
		if invite == nickname {
			return true
		}
	}
	return false
}

// String returns a string representation of the Channel
func (c *Channel) String() string {
	return fmt.Sprintf("Channel{Name: %s, Users: %d, Topic: %s}", c.Name, len(c.Users), c.Topic)
}

// GetUserList returns a list of usernames in the channel, with @ for operators
func (c *Channel) GetUserList() []string {
	userList := make([]string, 0, len(c.Users))
	for nickname, user := range c.Users {
		if c.Operators[nickname] {
			userList = append(userList, "@"+user.Nickname)
		} else {
			userList = append(userList, user.Nickname)
		}
	}
	return userList
}
