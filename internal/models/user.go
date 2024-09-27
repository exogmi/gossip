package models

import (
	"fmt"
	"sync"
	"time"
)

// ClientSession is a forward declaration to avoid circular imports
type ClientSession interface {
	SendMessage(message string) error
}

// User represents an IRC user
type User struct {
	ID              string
	Nickname        string
	Username        string
	Realname        string
	Host            string
	CreatedAt       time.Time
	LastSeen        time.Time
	LastDisconnect  time.Time
	Channels        []string
	Modes           UserModes
	ClientSessions  map[string]ClientSession
	sessionMutex    sync.RWMutex
}

// UserModes represents the modes a user can have
type UserModes struct {
	Away      bool
	Invisible bool
	Operator  bool
}

// NewUser creates a new User instance
func NewUser(nickname, username, realname, host string) *User {
	return &User{
		ID:              generateUniqueID(),
		Nickname:        nickname,
		Username:        username,
		Realname:        realname,
		Host:            host,
		CreatedAt:       time.Now(),
		LastSeen:        time.Now(),
		Channels:        make([]string, 0),
		Modes:           UserModes{},
		ClientSessions:  make(map[string]ClientSession),
	}
}

// AddClientSession adds a new client session for the user
func (u *User) AddClientSession(sessionID string, session ClientSession) {
	u.sessionMutex.Lock()
	defer u.sessionMutex.Unlock()
	u.ClientSessions[sessionID] = session
}

// RemoveClientSession removes a client session for the user
func (u *User) RemoveClientSession(sessionID string) {
	u.sessionMutex.Lock()
	defer u.sessionMutex.Unlock()
	delete(u.ClientSessions, sessionID)
	if len(u.ClientSessions) == 0 {
		u.LastDisconnect = time.Now()
	}
}

// BroadcastToSessions sends a message to all active sessions of the user
func (u *User) BroadcastToSessions(message string) {
	u.sessionMutex.RLock()
	defer u.sessionMutex.RUnlock()
	for _, session := range u.ClientSessions {
		session.SendMessage(message)
	}
}

// SetMode sets a mode for the user
func (u *User) SetMode(mode string, value bool) error {
	switch mode {
	case "a": // away
		u.Modes.Away = value
	case "i": // invisible
		u.Modes.Invisible = value
	case "o": // operator
		u.Modes.Operator = value
	default:
		return fmt.Errorf("unknown mode: %s", mode)
	}
	return nil
}

// IsInChannel checks if the user is in a specific channel
func (u *User) IsInChannel(channelName string) bool {
	for _, ch := range u.Channels {
		if ch == channelName {
			return true
		}
	}
	return false
}

// JoinChannel adds a channel to the user's list of channels
func (u *User) JoinChannel(channelName string) {
	if !u.IsInChannel(channelName) {
		u.Channels = append(u.Channels, channelName)
	}
}

// LeaveChannel removes a channel from the user's list of channels
func (u *User) LeaveChannel(channelName string) {
	for i, ch := range u.Channels {
		if ch == channelName {
			u.Channels = append(u.Channels[:i], u.Channels[i+1:]...)
			break
		}
	}
}

// UpdateLastSeen updates the user's last seen timestamp
func (u *User) UpdateLastSeen() {
	u.LastSeen = time.Now()
}

// String returns a string representation of the User
func (u *User) String() string {
	return fmt.Sprintf("User{ID: %s, Nickname: %s, Username: %s, Host: %s}", u.ID, u.Nickname, u.Username, u.Host)
}
