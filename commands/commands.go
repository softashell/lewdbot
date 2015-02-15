package commands

import (
	"fmt"
	"github.com/Philipp15b/go-steam/steamid"
	"github.com/softashell/lewdbot/regex"
	. "github.com/softashell/lewdbot/settings"
	"strings"
)

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

		if _, err := steamid.NewId(arg[1]); err != nil {
			return true, "ERROR: invalid steam id"
		}

		return true, fmt.Sprintf("added %s to admin list (but not really, that's not implemented)", arg[1])

	case "blacklist.add":
		arg := regex.BlacklistAddArguments.FindStringSubmatch(message)

		if len(arg) < 1 {
			return true, "not enough arguments"
		}

		id, err := steamid.NewId(arg[1])
		if err != nil {
			return true, "ERROR: invalid group id"
		}

		settings.SetGroupBlacklisted(id, true)
		return true, fmt.Sprintf("adding %s to group blacklist", arg[1])

	case "blacklist.remove":
		arg := regex.BlacklistRemoveArguments.FindStringSubmatch(message)

		if len(arg) < 1 {
			return true, "not enough arguments"
		}

		id, err := steamid.NewId(arg[1])
		if err != nil {
			return true, "ERROR: invalid group id"
		}

		settings.SetGroupBlacklisted(id, false)
		return true, fmt.Sprintf("removing %s from group blacklist", arg[1])

	default:
		return true, fmt.Sprintf("unknown command: %s", command)
	}

	return true, ""
}
