package shared

type Network interface {
	Main()
	ValidateChat(string) (bool, string)
	ValidateUser(string) (bool, string)
	ListChats() []string
	JoinChat(chat string)
	LeaveChat(chat string)
	ListAutojoinChats() []string
	ChatBlacklistAdd(chat string)
	ChatBlacklistRemove(chat string)
	ChatBlacklistList() []string
	MasterAdd(user string)
	MasterRemove(user string)
	MasterList() []string
}

type ReplyGenerator func(string) string
type MessageCleaner func(string) string
