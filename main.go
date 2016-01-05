package main

import (
	"encoding/json"
	"github.com/softashell/lewdbot/regex"
	. "github.com/softashell/lewdbot/settings"
	"github.com/softashell/lewdbot/steam"
	"log"
	"os"
	"strings"
)

type Configuration struct {
	Username string
	Password string
	Master   uint64
}

var configuration Configuration
var settings *Settings

func cleanMessage(message string) string {
	message = regex.Link.ReplaceAllString(message, "")
	message = regex.Emoticon.ReplaceAllString(message, "")
	message = regex.Junk.ReplaceAllString(message, "")
	message = regex.WikipediaCitations.ReplaceAllString(message, "")
	message = regex.Actions.ReplaceAllString(message, " ")
	message = regex.Russian.ReplaceAllString(message, "")
	message = regex.RepeatedWhitespace.ReplaceAllString(message, " ")

	return strings.TrimSpace(message)
}

func main() {
	file, err := os.Open("./config.json")
	if err != nil {
		log.Fatal(err)
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&configuration); err != nil {
		log.Fatal(err)
	}

	os.Mkdir("./data", 0777)
	os.Mkdir("./logs", 0777)

	init_chat()

	settings = LoadSettings("data/lewdbot.db")
	defer settings.Close()

	client := steam.NewClient(
		settings,
		configuration.Username,
		configuration.Password,
		configuration.Master,
		generateReply,
		cleanMessage,
	)
	client.Main()
}
