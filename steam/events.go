package steam

import (
	"fmt"
	"github.com/Philipp15b/go-steam"
	. "github.com/Philipp15b/go-steam/protocol/steamlang"
	"github.com/Philipp15b/go-steam/socialcache"
	"github.com/Philipp15b/go-steam/steamid"
	"github.com/softashell/lewdbot/commands"
	"github.com/softashell/lewdbot/regex"
	"log"
	"os"
	"strings"
)

func (c *Client) chatMsgEvent(e *steam.ChatMsgEvent) {
	if !e.IsMessage() {
		return
	}

	if c.Settings.IsUserBanned(e.ChatterId) {
		return
	}

	if c.isMaster(e.ChatterId) {
		if master, replies := commands.Handle(c, e.Message); master == true {
			for _, reply := range replies {
				c.client.Social.SendMessage(e.ChatterId, EChatEntryType_ChatMsg, reply)
			}
			if len(replies) == 0 {
				c.client.Social.SendMessage(e.ChatterId, EChatEntryType_ChatMsg, "I got nothing!")
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
				c.client.Social.SendMessage(e.ChatRoomId, EChatEntryType_ChatMsg, "Got it!")
				return
			case strings.HasSuffix(e.Message, "you may speak freely."):
				c.Settings.SetGroupQuiet(e.ChatRoomId, false)
				c.client.Social.SendMessage(e.ChatRoomId, EChatEntryType_ChatMsg, "Got it!")
				return
			case strings.HasSuffix(e.Message, "you can come here any time you'd like."):
				c.Settings.SetGroupAutojoin(e.ChatRoomId, true)
				c.client.Social.SendMessage(e.ChatRoomId, EChatEntryType_ChatMsg, "Got it!")
				return
			case strings.HasSuffix(e.Message, "stop coming here."):
				c.Settings.SetGroupAutojoin(e.ChatRoomId, false)
				c.client.Social.SendMessage(e.ChatRoomId, EChatEntryType_ChatMsg, "Got it!")
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
			c.client.Social.SendMessage(e.ChatterId, EChatEntryType_ChatMsg, "Иди нахуй")
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
			c.client.Social.SendMessage(e.ChatterId, EChatEntryType_ChatMsg, "Are you retarded?~")
		}
		return
	} else if regex.Greentext.MatchString(message) {
		c.client.Social.SendMessage(destination, EChatEntryType_ChatMsg, "Who are you quoting?~")
		return
	} else if regex.JustPunctuation.MatchString(message) || regex.LeadingNumbers.MatchString(message) {
		return
	}

	reply, learned := c.GenerateReply(message)
	reply = regex.Lewdbot.ReplaceAllString(reply, c.name(e.ChatterId))

	c.logMessage(destination, e.ChatterId, message, reply, learned)
	c.client.Social.SendMessage(destination, EChatEntryType_ChatMsg, reply)
}

// todo move to main
func (c *Client) logMessage(id steamid.SteamId, chatter steamid.SteamId, message string, reply string, learned bool) {
	var name string

	if learned { // If message was learned while generating reply add it to chatlog.txt and learn it every time
		f, err := os.OpenFile("./data/chatlog.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		if _, err = f.WriteString(fmt.Sprintf("%s\n\n", message)); err != nil {
			log.Print(err)
		}
	}

	if !c.isChatRoom(id) {
		name = c.name(id)
	} else {
		name = c.name(chatter)
	}

	filename := fmt.Sprintf("%d", id.ToUint64())
	text := fmt.Sprintf("%s: %s\nlewdbot: %s\n", name, message, reply)

	log.Printf("Learned: %t\n%s", learned, text)

	f, err := os.OpenFile(fmt.Sprintf("./logs/%s.txt", filename), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err = f.WriteString(text); err != nil {
		log.Print(err)
	}
}

func (c *Client) friendAddedEvent(e *steam.FriendAddedEvent) {
	c.client.Social.SendMessage(e.SteamId, EChatEntryType_ChatMsg, "Looking forward to working with you~ fu fu fu~")
}

func (c *Client) friendStateEvent(e *steam.FriendStateEvent) {
	switch e.Relationship {
	case EFriendRelationship_None:
		log.Printf("%s removed me from friends list", c.link(e.SteamId))
	case EFriendRelationship_RequestRecipient:
		log.Printf("%s added me to friends list", c.link(e.SteamId))

		if !c.Settings.IsUserBanned(e.SteamId) {
			c.client.Social.AddFriend(e.SteamId)
		} else {
			log.Printf("%s is banned, ignoring friend request", c.link(e.SteamId))
			c.client.Social.RemoveFriend(e.SteamId)
		}
	case EFriendRelationship_Friend:
		log.Printf("%s (%s) is now a friend", c.name(e.SteamId), c.link(e.SteamId))
		c.strangerList.Remove(e.SteamId)
	}
}

// Called after logging in and getting full friend and group list
func (c *Client) friendsListEvent(e *steam.FriendsListEvent) {
	// TODO: Clean up friends list when it's almost full
	// Accepts all pending friend invites
	for id, friend := range c.client.Social.Friends.GetCopy() {
		switch friend.Relationship {
		case EFriendRelationship_RequestInitiator:
			log.Printf("%s still hasn't accepted invite, consider removing", c.link(id))
		case EFriendRelationship_RequestRecipient:
			log.Printf("%s added me to friends list while I was offline", c.link(id))

			if !c.Settings.IsUserBanned(id) {
				c.client.Social.AddFriend(id)
			} else {
				log.Printf("%s is banned, ignoring friend request", c.link(id))
				c.client.Social.RemoveFriend(id)
			}
		}
	}

	// Lists pending group invites
	for id, group := range c.client.Social.Groups.GetCopy() {
		switch group.Relationship {
		case EClanRelationship_Invited:
			log.Printf("Pending invite to join group %s\n", c.link(id))

			//TODO:Actually accept them, needs some work on go-steam
		}
	}

	// Request missed chat messages when friends list is fully loaded
	c.client.Social.RequestOfflineMessages()
}

func (c *Client) chatInviteEvent(e *steam.ChatInviteEvent) {
	if e.ChatRoomType != EChatRoomType_Lobby { // Group chat or multi user chat
		log.Printf("Invited to %s (%d) by %s (%d)", e.ChatRoomName, e.ChatRoomId, c.name(e.PatronId), e.PatronId.ToUint64())

		if !c.Settings.IsGroupBlacklisted(e.ChatRoomId) {
			c.client.Social.SendMessage(e.PatronId, EChatEntryType_ChatMsg, "On my way~ I hope you will not keep me in your basement forever~")
			c.inviteList.Add(e.ChatRoomId, e.PatronId)
			c.client.Social.JoinChat(e.ChatRoomId)
		} else {
			log.Printf("User %s (%d) attempted to invite me to blacklisted group chat", c.name(e.PatronId), e.PatronId.ToUint64())
			c.client.Social.SendMessage(e.PatronId, EChatEntryType_ChatMsg, "Only disgusting nerds go there~")
		}
	} else { // Game lobby
		log.Printf("User %s (%d) attempted to invite me to game (%d) lobby", c.name(e.PatronId), e.PatronId.ToUint64(), e.GameId)
		c.client.Social.SendMessage(e.PatronId, EChatEntryType_ChatMsg, "I don't play bad video games~")
	}
}

func (c *Client) chatEnterEvent(e *steam.ChatEnterEvent) {
	inviter := c.inviteList.byId[e.ChatRoomId]

	if inviter != 0 {
		c.inviteList.Remove(e.ChatRoomId)
	}

	if e.EnterResponse == EChatRoomEnterResponse_Success {
		log.Printf("Joined %s (%s)", e.Name, e.ChatRoomId)
	} else {
		log.Printf("Failed to join %s! Respone: %s", e.ChatRoomId, e.EnterResponse)

		if inviter != 0 {
			switch e.EnterResponse {
			case EChatRoomEnterResponse_CommunityBan:
				c.client.Social.SendMessage(inviter, EChatEntryType_ChatMsg, "~banned from steam community~")
			case EChatRoomEnterResponse_Banned:
				c.client.Social.SendMessage(inviter, EChatEntryType_ChatMsg, "nerds don't want to see me there ;-;")
			default:
				c.client.Social.SendMessage(inviter, EChatEntryType_ChatMsg, "couldn't join chat room, better luck next time~")
			}
		}
	}
}

func (c *Client) chatMemberInfoEvent(e *steam.ChatMemberInfoEvent) {
	if e.Type == EChatInfoType_StateChange {
		if e.StateChangeInfo.ChatterActedOn == c.client.SteamId() {
			switch e.StateChangeInfo.StateChange {
			case EChatMemberStateChange_Left: // Doesn't get called
				log.Printf("Left room http://steamcommunity.com/gid/%d", e.ChatRoomId)
			case EChatMemberStateChange_Kicked:
				log.Printf("Kicked from %s by %s", e.ChatRoomId, c.name(e.StateChangeInfo.ChatterActedBy))
			case EChatMemberStateChange_Banned:
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

	if e.State == EPersonaState_Offline {
		return // Most likely a group update
	}

	c.strangerList.Add(
		socialcache.Friend{e.FriendId, e.Name, e.Avatar, EFriendRelationship_None,
			e.State, e.StateFlags, e.GameAppId, e.GameId, e.GameName})

	if e.StatusFlags&EClientPersonaStateFlag_PlayerName != 0 {
		c.strangerList.SetName(e.FriendId, e.Name)
	}
}

func (c *Client) autojoinGroups() {
	autojoin := c.Settings.ListGroupAutojoin()
	for _, group := range autojoin {
		c.client.Social.JoinChat(group)
	}
}
