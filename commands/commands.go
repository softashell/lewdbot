package commands

import (
	"fmt"
	"github.com/Philipp15b/go-steam/steamid"
	"github.com/softashell/lewdbot/regex"
	. "github.com/softashell/lewdbot/settings"
	"strings"
)

func blacklistAdd(settings Settings, arg1 string) []string {
	id, err := steamid.NewId(arg1)
	if err != nil {
		return []string{"invalid group id"}
	}

	settings.SetGroupBlacklisted(id, true)
	return []string{fmt.Sprintf("added %s to group blacklist", arg1)}
}

func blacklistRemove(settings Settings, arg1 string) []string {
	id, err := steamid.NewId(arg1)
	if err != nil {
		return []string{"invalid group id"}
	}

	settings.SetGroupBlacklisted(id, false)
	return []string{fmt.Sprintf("removed %s from group blacklist", arg1)}
}

func blacklistList(settings Settings) []string {
	groups := settings.ListGroupBlacklisted()
	var list []string
	for _, group := range groups {
		list = append(list, fmt.Sprintf("http://steamcommunity.com/gid/%s", group.String()))
	}
	return list
}

func masterAdd(settings Settings, arg1 string) []string {
	id, err := steamid.NewId(arg1)
	if err != nil {
		return []string{"invalid user id"}
	}

	settings.SetUserMaster(id, true)
	return []string{fmt.Sprintf("added %s to master list", arg1)}
}

func masterRemove(settings Settings, arg1 string) []string {
	id, err := steamid.NewId(arg1)
	if err != nil {
		return []string{"invalid user id"}
	}

	settings.SetUserMaster(id, false)
	return []string{fmt.Sprintf("removed %s from master list", arg1)}
}

func masterList(settings Settings) []string {
	users := settings.ListUserMaster()
	var list []string
	for _, user := range users {
		list = append(list, fmt.Sprintf("http://steamcommunity.com/profiles/%d", user.ToUint64()))
	}
	return list
}

// Handle takes the full command message and the settings struct and executes
// the command specified in the message. It returns a bool saying whether the
// regular response should be inhibited, and message(s) lewdbot should reply to
// the admin with.
func Handle(message string, settings Settings) (bool, []string) {
	if !strings.HasPrefix(message, "!") {
		return false, []string{}
	}

	command := regex.CommandName.FindStringSubmatch(message)[1]

	switch command {
	case "blacklist.add":
		arg := regex.BlacklistAddArguments.FindStringSubmatch(message)

		if len(arg) < 1 {
			return true, []string{"not enough arguments"}
		}

		return true, blacklistAdd(settings, arg[1])

	case "blacklist.remove":
		arg := regex.BlacklistRemoveArguments.FindStringSubmatch(message)

		if len(arg) < 1 {
			return true, []string{"not enough arguments"}
		}

		return true, blacklistRemove(settings, arg[1])

	case "blacklist.list":
		return true, blacklistList(settings)

	case "master.add":
		arg := regex.MasterAddArguments.FindStringSubmatch(message)

		if len(arg) < 1 {
			return true, []string{"not enough arguments"}
		}

		return true, masterAdd(settings, arg[1])

	case "master.remove":
		arg := regex.MasterRemoveArguments.FindStringSubmatch(message)

		if len(arg) < 1 {
			return true, []string{"not enough arguments"}
		}

		return true, masterRemove(settings, arg[1])

	case "master.list":
		return true, masterList(settings)

	default:
		return true, []string{fmt.Sprintf("unknown command: %s", command)}
	}
}
