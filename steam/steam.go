package steam

import (
	"fmt"
	"github.com/Philipp15b/go-steam"
	"github.com/Philipp15b/go-steam/internal/steamlang"
	"github.com/Philipp15b/go-steam/socialcache"
	"github.com/Philipp15b/go-steam/steamid"
	"github.com/softashell/lewdbot/commands"
	"github.com/softashell/lewdbot/regex"
	"github.com/softashell/lewdbot/settings"
	"github.com/softashell/lewdbot/shared"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type Client struct {
	Settings      *settings.Settings
	Username      string
	Password      string
	Master        uint64
	GenerateReply shared.ReplyGenerator
	CleanMessage  shared.MessageCleaner
	strangerList  *socialcache.FriendsList
	client        *steam.Client
}

func NewClient(s *settings.Settings, u string, p string, m uint64, g shared.ReplyGenerator, c shared.MessageCleaner) *Client {
	return &Client{
		s, u, p, m, g, c,
		socialcache.NewFriendsList(),
		steam.NewClient(),
	}
}

func (c *Client) chatMsgEvent(e *steam.ChatMsgEvent) {
	if !e.IsMessage() {
		return
	}

	if c.isMaster(e.ChatterId) {
		if master, replies := commands.Handle(c.client, e.Message, c.Settings); master == true {
			for _, reply := range replies {
				c.client.Social.SendMessage(e.ChatterId, steamlang.EChatEntryType_ChatMsg, reply)
			}
			if len(replies) == 0 {
				c.client.Social.SendMessage(e.ChatterId, steamlang.EChatEntryType_ChatMsg, "I got nothing!")
			}
			return
		}
	}

	message := e.Message

	if c.isChatRoom(e.ChatRoomId) {
		if strings.HasPrefix(strings.ToLower(e.Message), "lewdbot, ") {
			switch {
			case strings.HasSuffix(e.Message, "don't speak unless spoken to."):
				c.Settings.SetGroupQuiet(e.ChatRoomId, true)
				c.client.Social.SendMessage(e.ChatRoomId, steamlang.EChatEntryType_ChatMsg, "Got it!")
				return
			case strings.HasSuffix(e.Message, "you may speak freely."):
				c.Settings.SetGroupQuiet(e.ChatRoomId, false)
				c.client.Social.SendMessage(e.ChatRoomId, steamlang.EChatEntryType_ChatMsg, "Got it!")
				return
			case strings.HasSuffix(e.Message, "you can come here any time you'd like."):
				c.Settings.SetGroupAutojoin(e.ChatRoomId, true)
				c.client.Social.SendMessage(e.ChatRoomId, steamlang.EChatEntryType_ChatMsg, "Got it!")
				return
			case strings.HasSuffix(e.Message, "stop coming here."):
				c.Settings.SetGroupAutojoin(e.ChatRoomId, false)
				c.client.Social.SendMessage(e.ChatRoomId, steamlang.EChatEntryType_ChatMsg, "Got it!")
				return
			default:
				message = message[9:]
			}
		} else {
			if c.Settings.IsGroupQuiet(e.ChatRoomId) {
				// todo: logmessage here, without a reply
				return
			}
		}
	}

	if isRussian(e.Message) { // Should be called before cleanMessage since it replaces russian
		if !c.isChatRoom(e.ChatRoomId) { // Get out of here stalker
			c.client.Social.SendMessage(e.ChatterId, steamlang.EChatEntryType_ChatMsg, "Иди нахуй")
		}
		return
	}

	message = c.CleanMessage(message)
	var destination steamid.SteamId
	if c.isChatRoom(e.ChatRoomId) {
		destination = e.ChatRoomId
	} else {
		destination = e.ChatterId
	}

	if len(regex.NotActualText.ReplaceAllString(message, "")) < 3 { // Not enough actual text to bother replying
		if !c.isChatRoom(e.ChatRoomId) {
			c.client.Social.SendMessage(e.ChatterId, steamlang.EChatEntryType_ChatMsg, "Are you retarded?~")
		}
		return
	} else if regex.Greentext.MatchString(message) {
		c.client.Social.SendMessage(destination, steamlang.EChatEntryType_ChatMsg, "Who are you quoting?~")
		return
	} else if regex.JustPunctuation.MatchString(message) {
		return
	}

	reply := c.GenerateReply(message)
	reply = strings.Replace(reply, c.client.Social.GetPersonaName(), c.name(e.ChatterId), 1)

	c.logMessage(destination, e.ChatterId, message, reply)
	c.client.Social.SendMessage(destination, steamlang.EChatEntryType_ChatMsg, reply)
}

// todo move to main
func (c *Client) logMessage(id steamid.SteamId, chatter steamid.SteamId, message string, reply string) {
	var name string
	if !c.isChatRoom(id) {
		name = c.name(id)
	} else {
		name = c.name(chatter)
	}

	filename := fmt.Sprintf("%d", id.ToUint64())
	text := fmt.Sprintf("%s: %s\nlewdbot: %s\n", name, message, reply)

	log.Print(text)

	f, err := os.OpenFile("chatlog.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err = f.WriteString(fmt.Sprintf("%s\n", message)); err != nil {
		log.Print(err)
	}

	f, err = os.OpenFile(fmt.Sprintf("./logs/%s.txt", filename), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err = f.WriteString(text); err != nil {
		log.Print(err)
	}
}

func (c *Client) friendStateEvent(e *steam.FriendStateEvent) {
	switch e.Relationship {
	case steamlang.EFriendRelationship_None:
		log.Printf("%s removed me from friends list", c.link(e.SteamId))
	case steamlang.EFriendRelationship_PendingInvitee:
		log.Printf("%s added me to friends list", c.link(e.SteamId))
		c.client.Social.AddFriend(e.SteamId)
	case steamlang.EFriendRelationship_Friend:
		log.Printf("%s (%s) is now a friend", c.name(e.SteamId), c.link(e.SteamId))
		c.strangerList.Remove(e.SteamId)
	}
}

// Adds friends who added bot while it was offline
func (c *Client) friendsListEvent(e *steam.FriendsListEvent) {
	for id, friend := range c.client.Social.Friends.GetCopy() {
		switch friend.Relationship {
		case steamlang.EFriendRelationship_RequestInitiator:
			log.Printf("%s still hasn't accepted invite, consider removing", c.link(id))
		case steamlang.EFriendRelationship_PendingInvitee:
			log.Printf("%s added me to friends list while I was offline", c.link(id))
			c.client.Social.AddFriend(id)
		}
	}
}

func (c *Client) chatInviteEvent(e *steam.ChatInviteEvent) {
	if e.ChatRoomType != steamlang.EChatRoomType_Lobby {
		log.Printf("Invited to %s (%s) by %s %d", e.ChatRoomName, e.ChatRoomId, c.name(e.PatronId), e.PatronId.ToUint64())

		if !c.Settings.IsGroupBlacklisted(e.ChatRoomId) {
			c.client.Social.SendMessage(e.PatronId, steamlang.EChatEntryType_ChatMsg, "On my way~ I hope you will not keep me in your basement forever~")
			c.client.Social.JoinChat(e.ChatRoomId)
		} else {
			log.Print("group is blacklisted ")
			c.client.Social.SendMessage(e.PatronId, steamlang.EChatEntryType_ChatMsg, "Only disgusting nerds go there~")
		}
	}
}

func (c *Client) chatEnterEvent(e *steam.ChatEnterEvent) {
	if e.EnterResponse == steamlang.EChatRoomEnterResponse_Success {
		log.Printf("Joined %s (%s)", e.Name, e.ChatRoomId)
	} else {
		log.Printf("Failed to join %s! Respone: %s", e.ChatRoomId, e.EnterResponse)
	}
}

func (c *Client) chatMemberInfoEvent(e *steam.ChatMemberInfoEvent) {
	if e.Type == steamlang.EChatInfoType_StateChange {
		if e.StateChangeInfo.ChatterActedOn == c.client.SteamId() {
			switch e.StateChangeInfo.StateChange {
			case steamlang.EChatMemberStateChange_Left: // Doesn't get called
				log.Printf("Left room http://steamcommunity.com/gid/%d", e.ChatRoomId)
			case steamlang.EChatMemberStateChange_Kicked:
				log.Printf("Kicked from %s by %s", e.ChatRoomId, c.name(e.StateChangeInfo.ChatterActedBy))
			case steamlang.EChatMemberStateChange_Banned:
				log.Printf("Kicked and banned from %s by %s", e.ChatRoomId, c.name(e.StateChangeInfo.ChatterActedBy))
			}
		}
	}
}

func (c *Client) personaStateEvent(e *steam.PersonaStateEvent) {
	if e.FriendId == c.client.SteamId() {
		return // Updating own status
	}

	_, err := c.client.Social.Friends.ById(e.FriendId)
	if err == nil {
		return // Is a friend already, no need to update manually
	}

	if e.State == steamlang.EPersonaState_Offline {
		return // Most likely a group update
	}

	c.strangerList.Add(
		socialcache.Friend{e.FriendId, e.Name, e.Avatar, steamlang.EFriendRelationship_None,
			e.State, e.StateFlags, e.GameAppId, e.GameId, e.GameName})

	if e.StatusFlags&steamlang.EClientPersonaStateFlag_PlayerName != 0 {
		c.strangerList.SetName(e.FriendId, e.Name)
	}
}

func (c *Client) autojoinGroups() {
	autojoin := c.Settings.ListGroupAutojoin()
	for _, group := range autojoin {
		c.client.Social.JoinChat(group)
	}
}

func (c *Client) Main() {
	myLoginInfo := new(steam.LogOnDetails)
	myLoginInfo.Username = c.Username
	myLoginInfo.Password = c.Password

	c.client.Connect()
	//defer client.Disconnect()

	for event := range c.client.Events() {
		switch e := event.(type) {
		case *steam.ConnectedEvent:
			log.Print("Connecting")
			c.client.Auth.LogOn(myLoginInfo)
		case *steam.MachineAuthUpdateEvent:
			ioutil.WriteFile("sentry", e.Hash, 0666)
		case *steam.LoggedOnEvent:
			log.Print("Logged on")
			c.client.Social.SetPersonaState(steamlang.EPersonaState_Online)
			go c.autojoinGroups()
		case *steam.DisconnectedEvent:
			log.Print("DisconnectedEvent: ", e)
			log.Print("attempting to reconnect")
			c.client.Connect()
		case *steam.ChatMsgEvent:
			go c.chatMsgEvent(e)
		case *steam.FriendStateEvent:
			go c.friendStateEvent(e)
		case *steam.FriendsListEvent:
			go c.friendsListEvent(e)
		case *steam.ChatInviteEvent:
			go c.chatInviteEvent(e)
		case *steam.ChatEnterEvent:
			go c.chatEnterEvent(e)
		case *steam.ChatMemberInfoEvent:
			go c.chatMemberInfoEvent(e)
		case *steam.FriendAddedEvent:
			c.client.Social.SendMessage(e.SteamId, steamlang.EChatEntryType_ChatMsg, "Looking forward to working with you~ fu fu fu~")
		case *steam.PersonaStateEvent:
			go c.personaStateEvent(e)
		case steam.FatalErrorEvent:
			log.Print("FatalErrorEvent: ", e)
		case error:
			log.Print("error: ", e)
		}
	}
}
