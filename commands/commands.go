package commands

import (
	"fmt"
	"github.com/softashell/lewdbot/regex"
	"github.com/softashell/lewdbot/shared"
	"strings"
)

var ( // autism
	invalidChatIdentifier = "invalid chat identifier '%s'"
	invalidUserIdentifier = "invalid user identifier '%s'"
)

func autojoinList(client shared.Network) []string {
	return client.ListAutojoinChats()
}

func blacklistAdd(client shared.Network, arg1 string) []string {
	valid, chat := client.ValidateChat(arg1)
	if !valid {
		return []string{fmt.Sprintf(invalidChatIdentifier, arg1)}
	}
	client.ChatBlacklistAdd(arg1)
	return []string{fmt.Sprintf("blacklisted %s", chat)}
}

func blacklistRemove(client shared.Network, arg1 string) []string {
	valid, chat := client.ValidateChat(arg1)
	if !valid {
		return []string{fmt.Sprintf(invalidChatIdentifier, arg1)}
	}
	client.ChatBlacklistRemove(arg1)
	return []string{fmt.Sprintf("removed %s from blacklist", chat)}
}

func blacklistList(client shared.Network) []string {
	return client.ChatBlacklistList()
}

func chatList(client shared.Network) []string {
	return client.ListChats()
}

func chatJoin(client shared.Network, arg1 string) []string {
	valid, chat := client.ValidateChat(arg1)
	if !valid {
		return []string{fmt.Sprintf(invalidChatIdentifier, arg1)}
	}
	client.JoinChat(arg1)
	return []string{fmt.Sprintf("joining chat %s", chat)}
}

func chatLeave(client shared.Network, arg1 string) []string {
	valid, chat := client.ValidateChat(arg1)
	if !valid {
		return []string{fmt.Sprintf(invalidChatIdentifier, arg1)}
	}
	client.LeaveChat(arg1)
	return []string{fmt.Sprintf("leaving chat %s", chat)}
}

func masterAdd(client shared.Network, arg1 string) []string {
	valid, user := client.ValidateUser(arg1)
	if !valid {
		return []string{fmt.Sprintf(invalidUserIdentifier, arg1)}
	}
	client.MasterAdd(arg1)
	return []string{fmt.Sprintf("adding master %s", user)}
}

func masterRemove(client shared.Network, arg1 string) []string {
	valid, user := client.ValidateUser(arg1)
	if !valid {
		return []string{fmt.Sprintf(invalidUserIdentifier, arg1)}
	}
	client.MasterRemove(arg1)
	return []string{fmt.Sprintf("removing master %s", user)}
}

func masterList(client shared.Network) []string {
	return client.MasterList()
}

func banAdd(client shared.Network, arg1 string) []string {
	valid, user := client.ValidateUser(arg1)
	if !valid {
		return []string{fmt.Sprintf(invalidUserIdentifier, arg1)}
	}
	client.BanAdd(arg1)
	return []string{fmt.Sprintf("banning user %s", user)}
}

func banRemove(client shared.Network, arg1 string) []string {
	valid, user := client.ValidateUser(arg1)
	if !valid {
		return []string{fmt.Sprintf(invalidUserIdentifier, arg1)}
	}
	client.BanRemove(arg1)
	return []string{fmt.Sprintf("unbanning user %s", user)}
}

func banList(client shared.Network) []string {
	return client.BanList()
}

// Handle takes the full command message and the settings struct and executes
// the command specified in the message. It returns a bool saying whether the
// regular response should be inhibited, and message(s) lewdbot should reply to
// the admin with.
func Handle(client shared.Network, message string) (bool, []string) {
	if !strings.HasPrefix(message, "!") || len(message) == 1 {
		return false, []string{}
	}

	command := regex.CommandName.FindStringSubmatch(message)[1]

	switch command {
	case "autojoin.list":
		return true, autojoinList(client)

	case "blacklist.add":
		arg := regex.BlacklistAddArguments.FindStringSubmatch(message)

		if len(arg) < 1 {
			return true, []string{"not enough arguments"}
		}

		return true, blacklistAdd(client, arg[1])

	case "blacklist.remove":
		arg := regex.BlacklistRemoveArguments.FindStringSubmatch(message)

		if len(arg) < 1 {
			return true, []string{"not enough arguments"}
		}

		return true, blacklistRemove(client, arg[1])

	case "blacklist.list":
		return true, blacklistList(client)

	case "chat.list":
		return true, chatList(client)

	case "chat.join":
		arg := regex.MasterAddArguments.FindStringSubmatch(message)

		if len(arg) < 1 {
			return true, []string{"not enough arguments"}
		}

		return true, chatJoin(client, arg[1])

	case "chat.leave":
		arg := regex.MasterAddArguments.FindStringSubmatch(message)

		if len(arg) < 1 {
			return true, []string{"not enough arguments"}
		}

		return true, chatLeave(client, arg[1])

	case "master.add":
		arg := regex.MasterAddArguments.FindStringSubmatch(message)

		if len(arg) < 1 {
			return true, []string{"not enough arguments"}
		}

		return true, masterAdd(client, arg[1])

	case "master.remove":
		arg := regex.MasterRemoveArguments.FindStringSubmatch(message)

		if len(arg) < 1 {
			return true, []string{"not enough arguments"}
		}

		return true, masterRemove(client, arg[1])

	case "master.list":
		return true, masterList(client)

	case "ban.add":
		arg := regex.MasterAddArguments.FindStringSubmatch(message)

		if len(arg) < 1 {
			return true, []string{"not enough arguments"}
		}

		return true, banAdd(client, arg[1])

	case "ban.remove":
		arg := regex.MasterRemoveArguments.FindStringSubmatch(message)

		if len(arg) < 1 {
			return true, []string{"not enough arguments"}
		}

		return true, banRemove(client, arg[1])

	case "ban.list":
		return true, banList(client)

	default:
		return true, []string{fmt.Sprintf("unknown command: %s", command)}
	}
}
