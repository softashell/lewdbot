package commands

import (
	"fmt"
	"github.com/Philipp15b/go-steam/steamid"
	"github.com/softashell/lewdbot/regex"
	. "github.com/softashell/lewdbot/settings"
	"strings"
)

func adminAdd(settings Settings, arg1 string) []string {
	if _, err := steamid.NewId(arg1); err != nil {
		return []string{"invalid steam id"}
	}

	return []string{fmt.Sprintf("added %s to admin list (but not really, that's not implemented)", arg1)}
}

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
		list = append(list, group.String())
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
	case "admin.add":
		arg := regex.AdminAddArguments.FindStringSubmatch(message)

		if len(arg) < 1 {
			return true, []string{"not enough arguments"}
		}

		return true, adminAdd(settings, arg[1])

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

	default:
		return true, []string{fmt.Sprintf("unknown command: %s", command)}
	}
}
