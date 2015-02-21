package main

import (
	"fmt"
	"github.com/Philipp15b/go-steam"
	"github.com/Philipp15b/go-steam/steamid"
	"github.com/softashell/lewdbot/regex"
	"strings"
)

func isRussian(message string) bool {
	if regex.Russian.MatchString(message) {
		return true
	}

	return false
}

func isChatRoom(steamid steamid.SteamId) bool {
	switch steamid.GetAccountType() {
	case 7: //steamlang.EAccountType_Clan:
		fmt.Print("clan")
		return true
	case 8: //steamlang.EAccountType_Chat:
		return true
	}

	return false
}

func steamLink(s steamid.SteamId) string {
	switch s.GetAccountType() {
	case 1: // EAccountType_Individual
		return fmt.Sprintf("https://steamcommunity.com/profiles/%d", s.ToUint64())
	case 7: // EAccountType_Clan
		return fmt.Sprintf("https://steamcommunity.com/gid/%d", s.ToUint64())
	}
	return s.ToString()
}

func steamName(client *steam.Client, friendid steamid.SteamId) string {
	nerd, err := client.Social.Friends.ById(friendid)
	if err == nil {
		return nerd.Name
	}

	nerd, err = StrangerList.ById(friendid)
	if err == nil {
		return nerd.Name
	}

	return "Nerdgin"
}

func cleanMessage(message string) string {
	message = regex.Link.ReplaceAllString(message, "")
	message = regex.Emoticon.ReplaceAllString(message, "")
	message = regex.Junk.ReplaceAllString(message, "")
	message = regex.WikipediaCitations.ReplaceAllString(message, "")
	message = regex.RepeatedWhitespace.ReplaceAllString(message, " ")

	// GET OUT OF HERE STALKER
	message = regex.Russian.ReplaceAllString(message, "")

	return strings.TrimSpace(message)
}

func isMaster(master steamid.SteamId) bool {
	if configuration.Master == master.ToUint64() {
		return true
	}

	if settings.IsUserMaster(master) {
		return true
	}

	return false
}
