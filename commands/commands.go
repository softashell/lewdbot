package commands

import (
	"fmt"
	"github.com/softashell/lewdbot/regex"
	"github.com/softashell/lewdbot/shared"
	"strings"
)

func autojoinList(client shared.Network) []string {
	return client.ListAutojoinChats()
}

func blacklistAdd(client shared.Network, arg1 string) []string {
	return []string{client.ChatBlacklistAdd(arg1)}
}

func blacklistRemove(client shared.Network, arg1 string) []string {
	return []string{client.ChatBlacklistRemove(arg1)}
}

func blacklistList(client shared.Network) []string {
	return client.ChatBlacklistList()
}

func chatList(client shared.Network) []string {
	return client.ListChats()
}

func chatJoin(client shared.Network, arg1 string) []string {
	return []string{client.JoinChat(arg1)}
}

func chatLeave(client shared.Network, arg1 string) []string {
	return []string{client.LeaveChat(arg1)}
}

func masterAdd(client shared.Network, arg1 string) []string {
	return []string{client.MasterAdd(arg1)}
}

func masterRemove(client shared.Network, arg1 string) []string {
	return []string{client.MasterRemove(arg1)}
}

func masterList(client shared.Network) []string {
	return client.MasterList()
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

	default:
		return true, []string{fmt.Sprintf("unknown command: %s", command)}
	}
}
