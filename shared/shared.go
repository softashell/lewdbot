package shared

type Network interface {
	Main()
}

type ReplyGenerator func(string) string
type MessageCleaner func(string) string
