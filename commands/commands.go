package commands

import (
	"fmt"
	"github.com/Philipp15b/go-steam/steamid"
	"github.com/softashell/lewdbot/regex"
	. "github.com/softashell/lewdbot/settings"
	"strings"
)

func adminAdd(settings Settings, arg1 string) string {
	if _, err := steamid.NewId(arg1); err != nil {
		return "invalid steam id"
	}

	return fmt.Sprintf("added %s to admin list (but not really, that's not implemented)", arg1)
}

func blacklistAdd(settings Settings, arg1 string) string {
	id, err := steamid.NewId(arg1)
	if err != nil {
		return "invalid group id"
	}

	settings.SetGroupBlacklisted(id, true)
	return fmt.Sprintf("added %s to group blacklist", arg1)
}

func blacklistRemove(settings Settings, arg1 string) string {
	id, err := steamid.NewId(arg1)
	if err != nil {
		return "invalid group id"
	}

	settings.SetGroupBlacklisted(id, false)
	return fmt.Sprintf("removed %s from group blacklist", arg1)
}

func Handle(message string, settings Settings) (bool, string) {
	if !strings.HasPrefix(message, "!") {
		return false, ""
	}

	command := regex.CommandName.FindStringSubmatch(message)[1]

	switch command {
	case "admin.add":
		arg := regex.AdminAddArguments.FindStringSubmatch(message)

		if len(arg) < 1 {
			return true, "not enough arguments"
		}

		return true, adminAdd(settings, arg[1])

	case "blacklist.add":
		arg := regex.BlacklistAddArguments.FindStringSubmatch(message)

		if len(arg) < 1 {
			return true, "not enough arguments"
		}

		return true, blacklistAdd(settings, arg[1])

	case "blacklist.remove":
		arg := regex.BlacklistRemoveArguments.FindStringSubmatch(message)

		if len(arg) < 1 {
			return true, "not enough arguments"
		}

		return true, blacklistRemove(settings, arg[1])

	default:
		return true, fmt.Sprintf("unknown command: %s", command)
	}
}
