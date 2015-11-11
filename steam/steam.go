package steam

import (
	"github.com/softashell/go-steam"
	"github.com/softashell/go-steam/_internal/steamlang"
	"github.com/softashell/go-steam/socialcache"
	"github.com/softashell/go-steam/steamid"
	"github.com/softashell/lewdbot/settings"
	"github.com/softashell/lewdbot/shared"
	"io/ioutil"
	"log"
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

func (c *Client) ValidateChat(group string) (bool, string) {
	id, err := steamid.NewId(group)
	if err != nil {
		return false, ""
	}
	return c.isChatRoom(id), c.link(id)
}

func (c *Client) ValidateUser(user string) (bool, string) {
	id, err := steamid.NewId(user)
	if err != nil {
		return false, ""
	}
	return !c.isChatRoom(id), c.link(id)
}

func (c *Client) ListChats() []string {
	var list []string
	for _, chat := range c.client.Social.Chats.GetCopy() {
		list = append(list, c.link(chat.GroupId))
	}
	return list
}

func (c *Client) JoinChat(group string) {
	id, err := steamid.NewId(group)
	if err == nil {
		c.client.Social.JoinChat(id)
	}
}

func (c *Client) LeaveChat(group string) {
	id, err := steamid.NewId(group)
	if err == nil {
		c.client.Social.LeaveChat(id)
	}
}

func (c *Client) ListAutojoinChats() []string {
	groups := c.Settings.ListGroupAutojoin()
	var list []string
	for _, group := range groups {
		list = append(list, c.link(group))
	}
	return list
}

func (c *Client) ChatBlacklistAdd(group string) {
	id, err := steamid.NewId(group)
	if err == nil {
		c.Settings.SetGroupBlacklisted(id, true)
	}
}

func (c *Client) ChatBlacklistRemove(group string) {
	id, err := steamid.NewId(group)
	if err == nil {
		c.Settings.SetGroupBlacklisted(id, false)
	}
}

func (c *Client) ChatBlacklistList() []string {
	groups := c.Settings.ListGroupBlacklisted()
	var list []string
	for _, group := range groups {
		list = append(list, c.link(group))
	}
	return list
}

func (c *Client) MasterAdd(user string) {
	id, err := steamid.NewId(user)
	if err == nil {
		c.Settings.SetUserMaster(id, true)
	}
}

func (c *Client) MasterRemove(user string) {
	id, err := steamid.NewId(user)
	if err == nil {
		c.Settings.SetUserMaster(id, false)
	}
}

func (c *Client) MasterList() []string {
	users := c.Settings.ListUserMaster()
	var list []string
	for _, user := range users {
		list = append(list, c.link(user))
	}
	return list
}

func (c *Client) Main() {
	myLoginInfo := new(steam.LogOnDetails)
	myLoginInfo.Username = c.Username
	myLoginInfo.Password = c.Password

	c.client.Connect()

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
