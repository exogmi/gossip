package protocol

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/exogmi/gossip/internal/models"
	"github.com/exogmi/gossip/internal/state"
)

type ProtocolHandler struct {
	stateManager *state.StateManager
	user         *models.User
}

func NewProtocolHandler(stateManager *state.StateManager) *ProtocolHandler {
	return &ProtocolHandler{
		stateManager: stateManager,
	}
}

func (ph *ProtocolHandler) HandleCommand(user *models.User, message *IRCMessage) ([]string, error) {
	if ph == nil {
		return nil, fmt.Errorf("ProtocolHandler is nil")
	}
	if message == nil {
		return nil, fmt.Errorf("IRCMessage is nil")
	}

	log.Printf("Handling command: %s", message.Command)

	switch message.Command {
	case "NICK":
		return ph.handleNickCommand(message.Params)
	case "USER":
		return ph.handleUserCommand(message.Params)
	case "JOIN":
		return ph.handleJoinCommand(user, message.Params)
	case "PART":
		return ph.handlePartCommand(user, message.Params)
	case "PRIVMSG":
		return ph.handlePrivmsgCommand(user, message.Params)
	case "QUIT":
		return ph.handleQuitCommand(user, message.Params)
	case "CAP":
		return ph.handleCapCommand(user, message.Params)
	case "PONG":
		return ph.handlePongCommand(user, message.Params)
	case "TOPIC":
		return ph.handleTopicCommand(user, message.Params)
	case "ISON":
		return ph.handleIsonCommand(user, message.Params)
	default:
		return nil, fmt.Errorf("unknown command: %s", message.Command)
	}
}

func (ph *ProtocolHandler) handlePongCommand(user *models.User, params []string) ([]string, error) {
	// PONG command doesn't require any action, just log it if needed
	log.Printf("Received PONG from user %s", user.Nickname)
	return nil, nil
}

func (ph *ProtocolHandler) handleTopicCommand(user *models.User, params []string) ([]string, error) {
	if len(params) < 1 {
		return []string{fmt.Sprintf(":%s 461 %s TOPIC :Not enough parameters", ph.stateManager.ServerName, user.Nickname)}, nil
	}

	channelName := params[0]
	channel, err := ph.stateManager.ChannelManager.GetChannel(channelName)
	if err != nil {
		return []string{fmt.Sprintf(":%s 403 %s %s :No such channel", ph.stateManager.ServerName, user.Nickname, channelName)}, nil
	}

	if len(params) == 1 {
		// User is requesting the current topic
		if channel.Topic == "" {
			return []string{fmt.Sprintf(":%s 331 %s %s :No topic is set", ph.stateManager.ServerName, user.Nickname, channelName)}, nil
		}
		return []string{fmt.Sprintf(":%s 332 %s %s :%s", ph.stateManager.ServerName, user.Nickname, channelName, channel.Topic)}, nil
	}

	// User is setting a new topic
	newTopic := strings.Join(params[1:], " ")
	if strings.HasPrefix(newTopic, ":") {
		newTopic = newTopic[1:]
	}
	channel.SetTopic(newTopic)

	// Broadcast the topic change to all users in the channel
	topicChangeMsg := fmt.Sprintf(":%s!%s@%s TOPIC %s :%s", user.Nickname, user.Username, user.Host, channelName, newTopic)
	ph.stateManager.ChannelManager.BroadcastToChannel(channel, &models.Message{
		Sender:  user,
		Content: topicChangeMsg,
		Type:    models.ServerMessage,
	}, nil)

	return []string{topicChangeMsg}, nil
}

func (ph *ProtocolHandler) handleIsonCommand(user *models.User, params []string) ([]string, error) {
	if len(params) < 1 {
		return []string{fmt.Sprintf(":%s 461 %s ISON :Not enough parameters", ph.stateManager.ServerName, user.Nickname)}, nil
	}

	onlineUsers := []string{}
	for _, nickname := range params {
		if ph.stateManager.UserManager.UserExists(nickname) {
			onlineUsers = append(onlineUsers, nickname)
		}
	}

	return []string{fmt.Sprintf(":%s 303 %s :%s", ph.stateManager.ServerName, user.Nickname, strings.Join(onlineUsers, " "))}, nil
}

func (ph *ProtocolHandler) GetUser() *models.User {
	return ph.user
}

func (ph *ProtocolHandler) handleNickCommand(params []string) ([]string, error) {
	if ph == nil || ph.stateManager == nil || ph.stateManager.UserManager == nil {
		return nil, fmt.Errorf("ProtocolHandler or its components are nil")
	}
	if len(params) < 1 {
		return []string{fmt.Sprintf(":%s %d %s :No nickname given", ph.stateManager.ServerName, ERR_NONICKNAMEGIVEN, ph.user.Nickname)}, nil
	}
	newNick := params[0]

	// Check if the new nickname is valid
	if !isValidNickname(newNick) {
		return []string{fmt.Sprintf(":%s %d %s %s :Erroneous nickname", ph.stateManager.ServerName, ERR_ERRONEUSNICKNAME, ph.user.Nickname, newNick)}, nil
	}

	// Check if the nickname is already in use
	existingUser, _ := ph.stateManager.UserManager.GetUser(newNick)
	if existingUser != nil && existingUser != ph.user {
		return []string{fmt.Sprintf(":%s %d %s %s :Nickname is already in use", ph.stateManager.ServerName, ERR_NICKNAMEINUSE, ph.user.Nickname, newNick)}, nil
	}

	if ph.user == nil {
		// Create a new user if it doesn't exist
		newUser := models.NewUser(newNick, "", "", "")
		if err := ph.stateManager.UserManager.AddUser(newUser); err != nil {
			log.Printf("Failed to add new user: %v", err)
			return nil, fmt.Errorf("failed to add new user: %w", err)
		}
		ph.user = newUser
		log.Printf("Created new user with nickname %s", newNick)
		return []string{fmt.Sprintf(":%s NICK %s", newNick, newNick)}, nil
	} else {
		oldNick := ph.user.Nickname
		if oldNick == newNick {
			// No change in nickname
			return nil, nil
		}
		log.Printf("Changing nickname for user %s to %s", oldNick, newNick)
		if err := ph.stateManager.UserManager.ChangeNickname(oldNick, newNick); err != nil {
			log.Printf("Failed to change nickname: %v", err)
			return nil, fmt.Errorf("failed to change nickname: %w", err)
		}
		ph.user.Nickname = newNick

		// Notify all channels the user is in about the nickname change
		nickChangeMsg := fmt.Sprintf(":%s!%s@%s NICK :%s", oldNick, ph.user.Username, ph.user.Host, newNick)
		serverMsg := &models.Message{
			Sender:  ph.user,
			Content: nickChangeMsg,
			Type:    models.ServerMessage,
		}
		for _, channelName := range ph.user.Channels {
			channel, err := ph.stateManager.ChannelManager.GetChannel(channelName)
			if err != nil {
				log.Printf("Failed to get channel %s: %v", channelName, err)
				continue
			}
			ph.stateManager.ChannelManager.BroadcastToChannel(channel, serverMsg, ph.user)
		}

		// Send the nickname change message to the user who changed their nickname
		return []string{nickChangeMsg}, nil
	}
}

func isValidNickname(nickname string) bool {
	// Implement nickname validation logic here
	// For example, you can use a regular expression to check if the nickname is valid
	// This is a simple example, you may want to adjust it based on your specific requirements
	match, _ := regexp.MatchString("^[a-zA-Z][a-zA-Z0-9_-]{0,8}$", nickname)
	return match
}

func (ph *ProtocolHandler) handleUserCommand(params []string) ([]string, error) {
	if len(params) < 4 {
		return nil, fmt.Errorf("not enough parameters for USER command")
	}
	username, _, _, realname := params[0], params[1], params[2], params[3]

	if ph.user == nil {
		return nil, fmt.Errorf("user not initialized")
	}

	log.Printf("Setting user information for %s: username=%s, realname=%s", ph.user.Nickname, username, realname)

	ph.user.Username = username
	ph.user.Realname = realname
	ph.user.Host = "localhost" // Set a default host

	if err := ph.stateManager.UserManager.UpdateUser(ph.user); err != nil {
		log.Printf("Failed to update user: %v", err)
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	welcomeMsg := []string{
		fmt.Sprintf(":%s 001 %s :Welcome to the Gossip IRC Network %s!%s@%s",
			ph.stateManager.ServerName, ph.user.Nickname, ph.user.Nickname, ph.user.Username, ph.user.Host),
		fmt.Sprintf(":%s 002 %s :Your host is %s, running version 1.0",
			ph.stateManager.ServerName, ph.user.Nickname, ph.stateManager.ServerName),
		fmt.Sprintf(":%s 003 %s :This server was created %s",
			ph.stateManager.ServerName, ph.user.Nickname, time.Now().Format(time.RFC1123)),
		fmt.Sprintf(":%s 004 %s %s 1.0 o o",
			ph.stateManager.ServerName, ph.user.Nickname, ph.stateManager.ServerName),
	}
	return welcomeMsg, nil
}

func (ph *ProtocolHandler) handleJoinCommand(user *models.User, params []string) ([]string, error) {
	if len(params) < 1 {
		return nil, fmt.Errorf("not enough parameters for JOIN command")
	}
	channelName := params[0]

	log.Printf("User %s is joining channel %s", user.Nickname, channelName)

	_, err := ph.stateManager.ChannelManager.GetChannel(channelName)
	if err != nil {
		log.Printf("Channel %s not found, creating new channel", channelName)
		_, err = ph.stateManager.ChannelManager.CreateChannel(channelName, user)
		if err != nil {
			log.Printf("Failed to create channel %s: %v", channelName, err)
			return nil, fmt.Errorf("failed to create channel: %w", err)
		}
	}

	if err := ph.stateManager.ChannelManager.JoinChannel(user, channelName); err != nil {
		log.Printf("Failed to join channel %s: %v", channelName, err)
		return nil, fmt.Errorf("failed to join channel: %w", err)
	}

	// The JoinChannel function now handles sending all necessary messages
	return nil, nil
}

func (ph *ProtocolHandler) handlePartCommand(user *models.User, params []string) ([]string, error) {
	if len(params) < 1 {
		return nil, fmt.Errorf("not enough parameters for PART command")
	}
	channelName := params[0]

	log.Printf("User %s is leaving channel %s", user.Nickname, channelName)

	if err := ph.stateManager.ChannelManager.LeaveChannel(user, channelName); err != nil {
		log.Printf("Failed to leave channel %s: %v", channelName, err)
		return nil, fmt.Errorf("failed to leave channel: %w", err)
	}

	partMsg := []string{fmt.Sprintf(":%s!%s@%s PART %s", user.Nickname, user.Username, user.Host, channelName)}
	return partMsg, nil
}

func (ph *ProtocolHandler) handlePrivmsgCommand(user *models.User, params []string) ([]string, error) {
	if len(params) < 2 {
		return nil, fmt.Errorf("not enough parameters for PRIVMSG command")
	}
	target, message := params[0], params[1]

	// Remove the leading colon from the message
	if strings.HasPrefix(message, ":") {
		message = message[1:]
	}

	log.Printf("User %s is sending a message to %s: %s", user.Nickname, target, message)

	if strings.HasPrefix(target, "#") {
		channel, err := ph.stateManager.ChannelManager.GetChannel(target)
		if err != nil {
			log.Printf("Channel %s not found", target)
			return nil, fmt.Errorf("channel not found: %s", target)
		}
		msg := models.NewMessage(user, target, message, models.ChannelMessage)
		ph.stateManager.MessageStore.StoreMessage(msg)
		ph.stateManager.ChannelManager.BroadcastToChannel(channel, msg, user)
	} else {
		targetUser, err := ph.stateManager.UserManager.GetUser(target)
		if err != nil {
			log.Printf("User %s not found", target)
			return nil, fmt.Errorf("user not found: %s", target)
		}
		msg := models.NewMessage(user, target, message, models.PrivateMessage)
		ph.stateManager.MessageStore.StoreMessage(msg)
		formattedMsg := fmt.Sprintf(":%s!%s@%s PRIVMSG %s :%s", user.Nickname, user.Username, user.Host, target, message)
		targetUser.BroadcastToSessions(formattedMsg)
	}

	return nil, nil
}

func (ph *ProtocolHandler) handleQuitCommand(user *models.User, params []string) ([]string, error) {
	quitMessage := "Quit"
	if len(params) > 0 {
		quitMessage = params[0]
	}

	log.Printf("User %s is quitting: %s", user.Nickname, quitMessage)

	// Remove user from all channels
	for _, channelName := range user.Channels {
		ph.stateManager.ChannelManager.LeaveChannel(user, channelName)
	}

	// Remove user from UserManager
	ph.stateManager.UserManager.RemoveUser(user.Nickname)

	quitMsg := []string{fmt.Sprintf(":%s!%s@%s QUIT :%s", user.Nickname, user.Username, user.Host, quitMessage)}
	return quitMsg, nil
}

func (ph *ProtocolHandler) handleCapCommand(user *models.User, params []string) ([]string, error) {
	if len(params) < 1 {
		return nil, fmt.Errorf("not enough parameters for CAP command")
	}

	subCommand := params[0]
	log.Printf("Handling CAP command for user: %s", subCommand)

	switch subCommand {
	case "LS":
		return []string{"CAP * LS :"}, nil
	case "REQ":
		return []string{"CAP * ACK :"}, nil
	case "END":
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown CAP subcommand: %s", subCommand)
	}
}
