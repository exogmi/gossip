package protocol

import (
	"fmt"
	"strings"

	"github.com/exogmi/gossip/internal/models"
	"github.com/exogmi/gossip/internal/state"
)

 type ProtocolHandler struct {
     stateManager *state.StateManager
 }

 func NewProtocolHandler(stateManager *state.StateManager) *ProtocolHandler {
     return &ProtocolHandler{
         stateManager: stateManager,
     }
 }

 func (ph *ProtocolHandler) HandleCommand(user *models.User, command *Command) (string, error) {
     switch command.Name {
     case "NICK":
         return ph.handleNickCommand(user, command.Params)
     case "USER":
         return ph.handleUserCommand(user, command.Params)
     case "JOIN":
         return ph.handleJoinCommand(user, command.Params)
     case "PART":
         return ph.handlePartCommand(user, command.Params)
     case "PRIVMSG":
         return ph.handlePrivmsgCommand(user, command.Params)
     case "QUIT":
         return ph.handleQuitCommand(user, command.Params)
     case "CAP":
         return ph.handleCapCommand(user, command.Params)
     default:
         return "", fmt.Errorf("unknown command: %s", command.Name)
     }
 }

 func (ph *ProtocolHandler) handleNickCommand(user *models.User, params []string) (string, error) {
     if len(params) < 1 {
         return "", fmt.Errorf("not enough parameters for NICK command")
     }
     newNick := params[0]

     if err := ph.stateManager.UserManager.ChangeNickname(user.Nickname, newNick); err != nil {
         return "", fmt.Errorf("failed to change nickname: %w", err)
     }

     user.Nickname = newNick
     return fmt.Sprintf(":%s NICK %s", user.Nickname, newNick), nil
 }

 func (ph *ProtocolHandler) handleUserCommand(user *models.User, params []string) (string, error) {
     if len(params) < 4 {
         return "", fmt.Errorf("not enough parameters for USER command")
     }
     username, _, _, realname := params[0], params[1], params[2], params[3]

     user.Username = username
     user.Realname = realname

     welcomeMsg := fmt.Sprintf(":%s 001 %s :Welcome to the Gossip IRC Network %s!%s@%s",
         ph.stateManager.ServerName, user.Nickname, user.Nickname, user.Username, user.Host)
     return welcomeMsg, nil
 }

 func (ph *ProtocolHandler) handleJoinCommand(user *models.User, params []string) (string, error) {
     if len(params) < 1 {
         return "", fmt.Errorf("not enough parameters for JOIN command")
     }
     channelName := params[0]

     channel, err := ph.stateManager.ChannelManager.GetChannel(channelName)
     if err != nil {
         channel, err = ph.stateManager.ChannelManager.CreateChannel(channelName, user)
         if err != nil {
             return "", fmt.Errorf("failed to create channel: %w", err)
         }
     }

     if err := ph.stateManager.ChannelManager.JoinChannel(user, channelName); err != nil {
         return "", fmt.Errorf("failed to join channel: %w", err)
     }

     joinMsg := fmt.Sprintf(":%s!%s@%s JOIN %s", user.Nickname, user.Username, user.Host, channelName)
     topicMsg := fmt.Sprintf(":%s 332 %s %s :%s", ph.stateManager.ServerName, user.Nickname, channelName, channel.Topic)
     nameReplyMsg := fmt.Sprintf(":%s 353 %s = %s :%s", ph.stateManager.ServerName, user.Nickname, channelName, strings.Join(channel.GetUserList(), " "))
     endOfNamesMsg := fmt.Sprintf(":%s 366 %s %s :End of /NAMES list", ph.stateManager.ServerName, user.Nickname, channelName)

     return fmt.Sprintf("%s\r\n%s\r\n%s\r\n%s", joinMsg, topicMsg, nameReplyMsg, endOfNamesMsg), nil
 }

 func (ph *ProtocolHandler) handlePartCommand(user *models.User, params []string) (string, error) {
     if len(params) < 1 {
         return "", fmt.Errorf("not enough parameters for PART command")
     }
     channelName := params[0]

     if err := ph.stateManager.ChannelManager.LeaveChannel(user, channelName); err != nil {
         return "", fmt.Errorf("failed to leave channel: %w", err)
     }

     partMsg := fmt.Sprintf(":%s!%s@%s PART %s", user.Nickname, user.Username, user.Host, channelName)
     return partMsg, nil
 }

 func (ph *ProtocolHandler) handlePrivmsgCommand(user *models.User, params []string) (string, error) {
     if len(params) < 2 {
         return "", fmt.Errorf("not enough parameters for PRIVMSG command")
     }
     target, message := params[0], strings.Join(params[1:], " ")

     if strings.HasPrefix(target, "#") {
         channel, err := ph.stateManager.ChannelManager.GetChannel(target)
         if err != nil {
             return "", fmt.Errorf("channel not found: %s", target)
         }
         msg := models.NewMessage(user, target, message, models.ChannelMessage)
         ph.stateManager.MessageStore.StoreMessage(msg)
         ph.stateManager.ChannelManager.BroadcastToChannel(channel, msg, user)
     } else {
         targetUser, err := ph.stateManager.UserManager.GetUser(target)
         if err != nil {
             return "", fmt.Errorf("user not found: %s", target)
         }
         msg := models.NewMessage(user, target, message, models.PrivateMessage)
         ph.stateManager.MessageStore.StoreMessage(msg)
         // TODO: Implement message delivery to target user
     }

     return "", nil
 }

 func (ph *ProtocolHandler) handleQuitCommand(user *models.User, params []string) (string, error) {
     quitMessage := "Quit"
     if len(params) > 0 {
         quitMessage = params[0]
     }

     // Remove user from all channels
     for _, channelName := range user.Channels {
         ph.stateManager.ChannelManager.LeaveChannel(user, channelName)
     }

     // Remove user from UserManager
     ph.stateManager.UserManager.RemoveUser(user.Nickname)

     quitMsg := fmt.Sprintf(":%s!%s@%s QUIT :%s", user.Nickname, user.Username, user.Host, quitMessage)
     return quitMsg, nil
 }

 func (ph *ProtocolHandler) handleCapCommand(user *models.User, params []string) (string, error) {
     if len(params) < 1 {
         return "", fmt.Errorf("not enough parameters for CAP command")
     }

     subCommand := params[0]
     switch subCommand {
     case "LS":
         return "CAP * LS :", nil
     case "REQ":
         return "CAP * ACK :", nil
     case "END":
         return "", nil
     default:
         return "", fmt.Errorf("unknown CAP subcommand: %s", subCommand)
     }
 }
