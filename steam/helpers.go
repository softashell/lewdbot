package steam

import (
	"fmt"
	"github.com/softashell/go-steam/steamid"
	"github.com/softashell/lewdbot/regex"
)

func (c *Client) isChatRoom(steamid steamid.SteamId) bool {
	switch steamid.GetAccountType() {
	case 7: //steamlang.EAccountType_Clan:
		fmt.Print("clan")
		return true
	case 8: //steamlang.EAccountType_Chat:
		return true
	}

	return false
}

func (c *Client) link(s steamid.SteamId) string {
	switch s.GetAccountType() {
	case 1: // EAccountType_Individual
		return fmt.Sprintf("https://steamcommunity.com/profiles/%d", s.ToUint64())
	case 7: // EAccountType_Clan
		return fmt.Sprintf("https://steamcommunity.com/gid/%d", s.ToUint64())
	}
	return s.ToString()
}

func (c *Client) name(s steamid.SteamId) string {
	nerd, err := c.client.Social.Friends.ById(s)
	if err == nil {
		return nerd.Name
	}

	nerd, err = c.strangerList.ById(s)
	if err == nil {
		return nerd.Name
	}

	return "Nerdgin"
}

func (c *Client) isMaster(s steamid.SteamId) bool {
	if c.Master == s.ToUint64() {
		return true
	}

	if c.Settings.IsUserMaster(s) {
		return true
	}

	return false
}

// todo move these two
func isRussian(message string) bool {
	if regex.Russian.MatchString(message) {
		return true
	}

	return false
}
