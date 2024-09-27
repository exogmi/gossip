package models

import (
	"testing"
	"time"
)

func TestNewUser(t *testing.T) {
	nickname := "testuser"
	username := "testusername"
	realname := "Test User"
	host := "test.host"

	user := NewUser(nickname, username, realname, host)

	if user.Nickname != nickname {
		t.Errorf("Expected nickname %s, got %s", nickname, user.Nickname)
	}
	if user.Username != username {
		t.Errorf("Expected username %s, got %s", username, user.Username)
	}
	if user.Realname != realname {
		t.Errorf("Expected realname %s, got %s", realname, user.Realname)
	}
	if user.Host != host {
		t.Errorf("Expected host %s, got %s", host, user.Host)
	}
	if user.ID == "" {
		t.Error("Expected non-empty ID")
	}
	if user.CreatedAt.IsZero() {
		t.Error("Expected non-zero CreatedAt")
	}
	if user.LastSeen.IsZero() {
		t.Error("Expected non-zero LastSeen")
	}
	if len(user.Channels) != 0 {
		t.Error("Expected empty Channels slice")
	}
}

func TestUserSetMode(t *testing.T) {
	user := NewUser("testuser", "testusername", "Test User", "test.host")

	tests := []struct {
		mode  string
		value bool
	}{
		{"a", true},
		{"i", true},
		{"o", true},
		{"a", false},
		{"i", false},
		{"o", false},
	}

	for _, tt := range tests {
		err := user.SetMode(tt.mode, tt.value)
		if err != nil {
			t.Errorf("SetMode(%s, %v) returned unexpected error: %v", tt.mode, tt.value, err)
		}

		var actual bool
		switch tt.mode {
		case "a":
			actual = user.Modes.Away
		case "i":
			actual = user.Modes.Invisible
		case "o":
			actual = user.Modes.Operator
		}

		if actual != tt.value {
			t.Errorf("SetMode(%s, %v) = %v, want %v", tt.mode, tt.value, actual, tt.value)
		}
	}

	// Test invalid mode
	err := user.SetMode("x", true)
	if err == nil {
		t.Error("SetMode with invalid mode should return an error")
	}
}

func TestUserChannelOperations(t *testing.T) {
	user := NewUser("testuser", "testusername", "Test User", "test.host")
	channel := "testchannel"

	// Test JoinChannel
	user.JoinChannel(channel)
	if !user.IsInChannel(channel) {
		t.Errorf("Expected user to be in channel %s", channel)
	}

	// Test joining the same channel again (should not duplicate)
	user.JoinChannel(channel)
	if len(user.Channels) != 1 {
		t.Errorf("Expected user to be in 1 channel, got %d", len(user.Channels))
	}

	// Test LeaveChannel
	user.LeaveChannel(channel)
	if user.IsInChannel(channel) {
		t.Errorf("Expected user to not be in channel %s", channel)
	}

	// Test leaving a channel the user is not in (should not error)
	user.LeaveChannel(channel)
}

func TestUserUpdateLastSeen(t *testing.T) {
	user := NewUser("testuser", "testusername", "Test User", "test.host")
	time.Sleep(time.Millisecond) // Ensure some time passes

	oldLastSeen := user.LastSeen
	user.UpdateLastSeen()

	if !user.LastSeen.After(oldLastSeen) {
		t.Error("Expected LastSeen to be updated")
	}
}

func TestNewChannel(t *testing.T) {
	name := "testchannel"
	channel := NewChannel(name)

	if channel.Name != name {
		t.Errorf("Expected channel name %s, got %s", name, channel.Name)
	}
	if channel.CreatedAt.IsZero() {
		t.Error("Expected non-zero CreatedAt")
	}
	if len(channel.Users) != 0 {
		t.Error("Expected empty Users map")
	}
	if channel.UserLimits != 0 {
		t.Error("Expected UserLimits to be 0")
	}
	if len(channel.BanList) != 0 {
		t.Error("Expected empty BanList")
	}
	if len(channel.InviteList) != 0 {
		t.Error("Expected empty InviteList")
	}
}

func TestChannelSetMode(t *testing.T) {
	channel := NewChannel("testchannel")

	tests := []struct {
		mode  string
		value bool
	}{
		{"i", true},
		{"m", true},
		{"n", true},
		{"p", true},
		{"s", true},
		{"t", true},
		{"i", false},
		{"m", false},
		{"n", false},
		{"p", false},
		{"s", false},
		{"t", false},
	}

	for _, tt := range tests {
		err := channel.SetMode(tt.mode, tt.value)
		if err != nil {
			t.Errorf("SetMode(%s, %v) returned unexpected error: %v", tt.mode, tt.value, err)
		}

		var actual bool
		switch tt.mode {
		case "i":
			actual = channel.Modes.InviteOnly
		case "m":
			actual = channel.Modes.Moderated
		case "n":
			actual = channel.Modes.NoExternal
		case "p":
			actual = channel.Modes.Private
		case "s":
			actual = channel.Modes.Secret
		case "t":
			actual = channel.Modes.TopicSettableOnlyByOps
		}

		if actual != tt.value {
			t.Errorf("SetMode(%s, %v) = %v, want %v", tt.mode, tt.value, actual, tt.value)
		}
	}

	// Test invalid mode
	err := channel.SetMode("x", true)
	if err == nil {
		t.Error("SetMode with invalid mode should return an error")
	}
}

func TestChannelUserOperations(t *testing.T) {
	channel := NewChannel("testchannel")
	user := NewUser("testuser", "testusername", "Test User", "test.host")

	// Test AddUser
	channel.AddUser(user)
	if _, exists := channel.Users[user.Nickname]; !exists {
		t.Errorf("Expected user %s to be in channel", user.Nickname)
	}

	// Test RemoveUser
	channel.RemoveUser(user.Nickname)
	if _, exists := channel.Users[user.Nickname]; exists {
		t.Errorf("Expected user %s to be removed from channel", user.Nickname)
	}

	// Test removing a user that's not in the channel (should not error)
	channel.RemoveUser(user.Nickname)
}

func TestChannelBanOperations(t *testing.T) {
	channel := NewChannel("testchannel")
	userMask := "testuser!testuser@test.host"

	// Test banning a user
	channel.BanList = append(channel.BanList, userMask)

	if !channel.IsBanned(userMask) {
		t.Errorf("Expected user mask %s to be banned", userMask)
	}

	// Test unbanning a user
	channel.BanList = []string{}

	if channel.IsBanned(userMask) {
		t.Errorf("Expected user mask %s to be unbanned", userMask)
	}
}

func TestChannelInviteOperations(t *testing.T) {
	channel := NewChannel("testchannel")
	nickname := "testuser"

	// Test inviting a user
	channel.InviteList = append(channel.InviteList, nickname)

	if !channel.IsInvited(nickname) {
		t.Errorf("Expected user %s to be invited", nickname)
	}

	// Test uninviting a user
	channel.InviteList = []string{}

	if channel.IsInvited(nickname) {
		t.Errorf("Expected user %s to be uninvited", nickname)
	}
}

func TestNewMessage(t *testing.T) {
	sender := NewUser("sender", "senderusername", "Sender User", "sender.host")
	target := "testtarget"
	content := "Test message content"
	msgType := PrivateMessage

	message := NewMessage(sender, target, content, msgType)

	if message.Sender != sender {
		t.Errorf("Expected sender %v, got %v", sender, message.Sender)
	}
	if message.Target != target {
		t.Errorf("Expected target %s, got %s", target, message.Target)
	}
	if message.Content != content {
		t.Errorf("Expected content %s, got %s", content, message.Content)
	}
	if message.Type != msgType {
		t.Errorf("Expected message type %v, got %v", msgType, message.Type)
	}
	if message.ID == "" {
		t.Error("Expected non-empty ID")
	}
	if message.Timestamp.IsZero() {
		t.Error("Expected non-zero Timestamp")
	}
}

func TestMessageTypeChecks(t *testing.T) {
	sender := NewUser("sender", "senderusername", "Sender User", "sender.host")
	target := "testtarget"
	content := "Test message content"

	privateMsg := NewMessage(sender, target, content, PrivateMessage)
	if !privateMsg.IsPrivate() {
		t.Error("Expected PrivateMessage to be private")
	}
	if privateMsg.IsChannelMessage() {
		t.Error("Expected PrivateMessage not to be a channel message")
	}

	channelMsg := NewMessage(sender, target, content, ChannelMessage)
	if channelMsg.IsPrivate() {
		t.Error("Expected ChannelMessage not to be private")
	}
	if !channelMsg.IsChannelMessage() {
		t.Error("Expected ChannelMessage to be a channel message")
	}
}

func TestMessageFormattedTimestamp(t *testing.T) {
	sender := NewUser("sender", "senderusername", "Sender User", "sender.host")
	message := NewMessage(sender, "target", "content", PrivateMessage)

	formatted := message.FormattedTimestamp()
	parsed, err := time.Parse(time.RFC3339, formatted)
	if err != nil {
		t.Errorf("Failed to parse formatted timestamp: %v", err)
	}

	if !parsed.Equal(message.Timestamp) {
		t.Errorf("Parsed timestamp %v does not match original timestamp %v", parsed, message.Timestamp)
	}
}
