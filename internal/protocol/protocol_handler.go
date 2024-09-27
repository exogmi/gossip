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
	case "MODE":
		return ph.handleModeCommand(user, message.Params)
	case "KICK":
		return ph.handleKickCommand(user, message.Params)
	case "BAN":
		return ph.handleBanCommand(user, message.Params)
	default:
		return nil, fmt.Errorf("unknown command: %s", message.Command)
	}
}

func (ph *ProtocolHandler) handleModeCommand(user *models.User, params []string) ([]string, error) {
	if len(params) < 1 {
		return []string{fmt.Sprintf(":%s 461 %s MODE :Not enough parameters", ph.stateManager.ServerName, user.Nickname)}, nil
	}

	targetName := params[0]
	channel, err := ph.stateManager.GetChannel(targetName)

	if err == nil { // Channel exists
		if len(params) < 2 {
			modes := "+"
			if channel.Key != "" {
				modes += "k"
				if user.IsInChannel(channel.Name) {
					modes += " " + channel.Key
				}
			}
			return []string{fmt.Sprintf(":%s 324 %s %s %s", ph.stateManager.ServerName, user.Nickname, targetName, modes)}, nil
		}

		flag := params[1]
		switch flag {
		case "+k", "-k":
			return ph.handleChannelKeyMode(user, channel, flag, params)
		case "+o", "-o", "+v", "-v":
			return ph.handleChannelUserMode(user, channel, flag, params)
		default:
			return []string{fmt.Sprintf(":%s 472 %s %s :Unknown MODE flag", ph.stateManager.ServerName, user.Nickname, flag)}, nil
		}
	} else if targetName == user.Nickname {
		if len(params) == 1 {
			return []string{fmt.Sprintf(":%s 221 %s +", ph.stateManager.ServerName, user.Nickname)}, nil
		}
		return []string{fmt.Sprintf(":%s 501 %s :Unknown MODE flag", ph.stateManager.ServerName, user.Nickname)}, nil
	} else {
		return []string{fmt.Sprintf(":%s 403 %s :No such channel", ph.stateManager.ServerName, targetName)}, nil
	}
}

func (ph *ProtocolHandler) handleChannelKeyMode(user *models.User, channel *models.Channel, flag string, params []string) ([]string, error) {
	if !user.IsInChannel(channel.Name) {
		return []string{fmt.Sprintf(":%s 442 %s :You're not on that channel", ph.stateManager.ServerName, channel.Name)}, nil
	}

	if flag == "+k" {
		if len(params) < 3 {
			return []string{fmt.Sprintf(":%s 461 %s MODE :Not enough parameters", ph.stateManager.ServerName, user.Nickname)}, nil
		}
		key := params[2]
		channel.Key = key
		msg := fmt.Sprintf(":%s!%s@%s MODE %s +k %s", user.Nickname, user.Username, user.Host, channel.Name, key)
		ph.stateManager.ChannelManager.BroadcastToChannel(channel, &models.Message{
			Sender:  user,
			Content: msg,
			Type:    models.ServerMessage,
		}, nil)
		return []string{msg}, nil
	} else { // -k
		channel.Key = ""
		msg := fmt.Sprintf(":%s!%s@%s MODE %s -k", user.Nickname, user.Username, user.Host, channel.Name)
		ph.stateManager.ChannelManager.BroadcastToChannel(channel, &models.Message{
			Sender:  user,
			Content: msg,
			Type:    models.ServerMessage,
		}, nil)
		return []string{msg}, nil
	}
}

func (ph *ProtocolHandler) handleChannelUserMode(user *models.User, channel *models.Channel, flag string, params []string) ([]string, error) {
	if len(params) < 3 {
		return []string{fmt.Sprintf(":%s 461 %s MODE :Not enough parameters", ph.stateManager.ServerName, user.Nickname)}, nil
	}

	if !user.IsInChannel(channel.Name) || !channel.Operators[user.Nickname] {
		return []string{fmt.Sprintf(":%s 482 %s %s :You're not channel operator", ph.stateManager.ServerName, user.Nickname, channel.Name)}, nil
	}

	targetUser := params[2]
	switch flag {
	case "+o", "-o":
		channel.Operators[targetUser] = (flag == "+o")
	case "+v", "-v":
		channel.Voices[targetUser] = (flag == "+v")
	}

	msg := fmt.Sprintf(":%s!%s@%s MODE %s %s %s", user.Nickname, user.Username, user.Host, channel.Name, flag, targetUser)
	ph.stateManager.ChannelManager.BroadcastToChannel(channel, &models.Message{
		Sender:  user,
		Content: msg,
		Type:    models.ServerMessage,
	}, nil)

	// Log mode change
	log.Printf("Mode change in channel %s: %s sets %s on %s", channel.Name, user.Nickname, flag, targetUser)

	// Update user list for all users in the channel
	userList := channel.GetUserList()
	for _, u := range channel.Users {
		u.BroadcastToSessions(fmt.Sprintf(":%s 353 %s = %s :%s", ph.stateManager.ServerName, u.Nickname, channel.Name, strings.Join(userList, " ")))
		u.BroadcastToSessions(fmt.Sprintf(":%s 366 %s %s :End of /NAMES list", ph.stateManager.ServerName, u.Nickname, channel.Name))
	}

	return []string{msg}, nil
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

func (ph *ProtocolHandler) handleKickCommand(user *models.User, params []string) ([]string, error) {
	if len(params) < 2 {
		return []string{fmt.Sprintf(":%s 461 %s KICK :Not enough parameters", ph.stateManager.ServerName, user.Nickname)}, nil
	}

	channelName, targetNick := params[0], params[1]
	reason := "No reason given"
	if len(params) > 2 {
		reason = strings.Join(params[2:], " ")
	}

	channel, err := ph.stateManager.GetChannel(channelName)
	if err != nil {
		return []string{fmt.Sprintf(":%s 403 %s %s :No such channel", ph.stateManager.ServerName, user.Nickname, channelName)}, nil
	}

	if !channel.Operators[user.Nickname] {
		return []string{fmt.Sprintf(":%s 482 %s %s :You're not channel operator", ph.stateManager.ServerName, user.Nickname, channelName)}, nil
	}

	targetUser, err := ph.stateManager.GetUser(targetNick)
	if err != nil {
		return []string{fmt.Sprintf(":%s 401 %s %s :No such nick", ph.stateManager.ServerName, user.Nickname, targetNick)}, nil
	}

	if !targetUser.IsInChannel(channelName) {
		return []string{fmt.Sprintf(":%s 441 %s %s %s :They aren't on that channel", ph.stateManager.ServerName, user.Nickname, targetNick, channelName)}, nil
	}

	if err := ph.stateManager.ChannelManager.LeaveChannel(targetUser, channelName); err != nil {
		return []string{fmt.Sprintf(":%s 491 %s %s :Could not kick user", ph.stateManager.ServerName, user.Nickname, targetNick)}, nil
	}
	kickMsg := fmt.Sprintf(":%s!%s@%s KICK %s %s :%s", user.Nickname, user.Username, user.Host, channelName, targetNick, reason)
	ph.stateManager.ChannelManager.BroadcastToChannel(channel, &models.Message{
		Sender:  user,
		Content: kickMsg,
		Type:    models.ServerMessage,
	}, nil)

	// Send the kick message to the kicked user
	targetUser.BroadcastToSessions(kickMsg)

	return []string{kickMsg}, nil
}

func (ph *ProtocolHandler) handleBanCommand(user *models.User, params []string) ([]string, error) {
	if len(params) < 2 {
		return []string{fmt.Sprintf(":%s 461 %s BAN :Not enough parameters", ph.stateManager.ServerName, user.Nickname)}, nil
	}

	channelName, targetMask := params[0], params[1]

	channel, err := ph.stateManager.GetChannel(channelName)
	if err != nil {
		return []string{fmt.Sprintf(":%s 403 %s %s :No such channel", ph.stateManager.ServerName, user.Nickname, channelName)}, nil
	}

	if !channel.Operators[user.Nickname] {
		return []string{fmt.Sprintf(":%s 482 %s %s :You're not channel operator", ph.stateManager.ServerName, user.Nickname, channelName)}, nil
	}

	channel.BanList = append(channel.BanList, targetMask)
	banMsg := fmt.Sprintf(":%s!%s@%s MODE %s +b %s", user.Nickname, user.Username, user.Host, channelName, targetMask)
	ph.stateManager.ChannelManager.BroadcastToChannel(channel, &models.Message{
		Sender:  user,
		Content: banMsg,
		Type:    models.ServerMessage,
	}, nil)

	log.Printf("User %s banned %s from channel %s", user.Nickname, targetMask, channelName)

	return []string{banMsg}, nil
}

func (ph *ProtocolHandler) GetUser() *models.User {
	return ph.user
}

func (ph *ProtocolHandler) handleNickCommand(params []string) ([]string, error) {
	if ph == nil || ph.stateManager == nil || ph.stateManager.UserManager == nil {
		return nil, fmt.Errorf("ProtocolHandler or its components are nil")
	}
	if len(params) < 1 {
		return []string{fmt.Sprintf(":%s %d * :No nickname given", ph.stateManager.ServerName, ERR_NONICKNAMEGIVEN)}, nil
	}
	newNick := params[0]

	// Check if the new nickname is valid
	if !isValidNickname(newNick) {
		return []string{fmt.Sprintf(":%s %d * %s :Erroneous nickname", ph.stateManager.ServerName, ERR_ERRONEUSNICKNAME, newNick)}, nil
	}

	// Check if the nickname is already in use
	existingUser, _ := ph.stateManager.UserManager.GetUser(newNick)
	if existingUser != nil {
		if ph.user == nil || existingUser != ph.user {
			return []string{fmt.Sprintf(":%s %d * %s :Nickname is already in use", ph.stateManager.ServerName, ERR_NICKNAMEINUSE, newNick)}, nil
		}
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
	key := ""
	if len(params) > 1 {
		key = params[1]
	}

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

	if err := ph.stateManager.ChannelManager.JoinChannel(user, channelName, key); err != nil {
		log.Printf("Failed to join channel %s: %v", channelName, err)
		if err.Error() == "cannot join channel: incorrect key" {
			return []string{fmt.Sprintf(":%s 475 %s %s :Cannot join channel (+k) - bad key", ph.stateManager.ServerName, user.Nickname, channelName)}, nil
		}
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

	// Dont' remove user from all channels
	//for _, channelName := range user.Channels {
	//	ph.stateManager.ChannelManager.LeaveChannel(user, channelName)
	// }

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
