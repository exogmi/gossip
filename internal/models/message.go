package models

import (
	"fmt"
	"time"
)

// MessageType represents the type of IRC message
type MessageType int

const (
	PrivateMessage MessageType = iota
	Notice
	ChannelMessage
	ServerMessage
)

// Message represents an IRC message
type Message struct {
	ID        string
	Sender    *User
	Target    string
	Content   string
	Timestamp time.Time
	Type      MessageType
}

// NewMessage creates a new Message instance
func NewMessage(sender *User, target, content string, msgType MessageType) *Message {
	return &Message{
		ID:        generateUniqueID(), // TODO: Implement this function
		Sender:    sender,
		Target:    target,
		Content:   content,
		Timestamp: time.Now(),
		Type:      msgType,
	}
}

// IsPrivate checks if the message is a private message
func (m *Message) IsPrivate() bool {
	return m.Type == PrivateMessage
}

// IsChannelMessage checks if the message is a channel message
func (m *Message) IsChannelMessage() bool {
	return m.Type == ChannelMessage
}

// FormattedTimestamp returns a formatted string of the message timestamp
func (m *Message) FormattedTimestamp() string {
	return m.Timestamp.Format(time.RFC3339Nano)
}

// String returns a string representation of the Message
func (m *Message) String() string {
	return fmt.Sprintf("Message{ID: %s, Sender: %s, Target: %s, Type: %d, Content: %s}", m.ID, m.Sender.Nickname, m.Target, m.Type, m.Content)
}
