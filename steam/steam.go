package steam

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/Philipp15b/go-steam"
	. "github.com/Philipp15b/go-steam/protocol/steamlang"
	"github.com/Philipp15b/go-steam/socialcache"
	"github.com/Philipp15b/go-steam/steamid"
	"github.com/softashell/lewdbot/settings"
	"github.com/softashell/lewdbot/shared"
)

type Client struct {
	Settings      *settings.Settings
	Username      string
	Password      string
	Master        uint64
	GenerateReply shared.ReplyGenerator
	CleanMessage  shared.MessageCleaner
	strangerList  *socialcache.FriendsList
	inviteList    *InviteList
	client        *steam.Client
}

func NewClient(s *settings.Settings, u string, p string, m uint64, g shared.ReplyGenerator, c shared.MessageCleaner) *Client {
	return &Client{
		s, u, p, m, g, c,
		socialcache.NewFriendsList(),
		NewInviteList(),
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

func (c *Client) BanAdd(user string) {
	id, err := steamid.NewId(user)
	if err == nil {
		c.Settings.SetUserBanned(id, true)

		c.client.Social.IgnoreFriend(id, true)
		c.client.Social.RemoveFriend(id)
	}
}

func (c *Client) BanRemove(user string) {
	id, err := steamid.NewId(user)
	if err == nil {
		c.Settings.SetUserBanned(id, false)

		c.client.Social.IgnoreFriend(id, false)
	}
}

func (c *Client) BanList() []string {
	users := c.Settings.ListUserBanned()
	var list []string
	for _, user := range users {
		list = append(list, c.link(user))
	}
	return list
}

func (c *Client) ShouldCleanUpFriends() bool {
	// Friend limit: 250 + (5 * level)

	return true
}

func (c *Client) RemoveDeadFriends() bool {
	log.Println("Removing dead friends")

	// TODO: Go trough all friends and find longest offline, if that's not enough compare last use time

	return true
}

func (c *Client) Main() error {
	myLoginInfo := new(steam.LogOnDetails)
	myLoginInfo.Username = c.Username
	myLoginInfo.Password = c.Password

	// Attempt to read existing auth hash to avoid steam guard
	myLoginInfo.SentryFileHash, _ = ioutil.ReadFile("./data/sentry")

	steam.InitializeSteamDirectory()

	c.client.Connect()

	for event := range c.client.Events() {
		switch e := event.(type) {
		case *steam.ConnectedEvent:
			log.Print("Connecting")
			c.client.Auth.LogOn(myLoginInfo)
		case *steam.MachineAuthUpdateEvent:
			log.Print("Updated auth hash, it should no longer ask for auth!")
			ioutil.WriteFile("./data/sentry", e.Hash, 0666)
		case *steam.LoggedOnEvent:
			log.Print("Logged on")
			c.client.Social.SetPersonaState(EPersonaState_Online)
			go c.autojoinGroups()
		case *steam.LogOnFailedEvent:
			if e.Result == EResult_AccountLogonDenied {
				log.Print("Steam guard isn't letting me in! Enter auth code:")
				var authcode string
				fmt.Scanf("%s", &authcode)
				myLoginInfo.AuthCode = authcode
			} else {
				// TODO: Handle EResult_InvalidLoginAuthCode
				return fmt.Errorf("LogOnFailedEvent: %s", e)
			}
		case *steam.DisconnectedEvent:
			log.Print("Disconnected")
			return fmt.Errorf("Disconnected from steam")
		case *steam.ChatMsgEvent:
			go c.chatMsgEvent(e)
		case *steam.FriendAddedEvent:
			go c.friendAddedEvent(e)
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
		case *steam.PersonaStateEvent:
			go c.personaStateEvent(e)
		case steam.FatalErrorEvent:
			return fmt.Errorf("FatalErrorEvent: %s", e)
		case error:
			log.Print("error: ", e)
		}
	}

	return nil
}
