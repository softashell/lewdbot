package shared

type Network interface {
	Main()
	ListChats() []string
	JoinChat(chat string) string
	LeaveChat(chat string) string
	ListAutojoinChats() []string
	ChatBlacklistAdd(chat string) string
	ChatBlacklistRemove(chat string) string
	ChatBlacklistList() []string
	MasterAdd(user string) string
	MasterRemove(user string) string
	MasterList() []string
}

type ReplyGenerator func(string) string
type MessageCleaner func(string) string
