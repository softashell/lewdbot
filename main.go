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
	message = regex.RepeatedWhitespace.ReplaceAllString(message, " ")

	// GET OUT OF HERE STALKER
	message = regex.Russian.ReplaceAllString(message, "")

	return strings.TrimSpace(message)
}

func main() {
	os.Mkdir("./data", 0777)
	os.Mkdir("./logs", 0777)

	file, _ := os.Open("./config.json")
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&configuration); err != nil {
		log.Fatal(err)
	}

	init_chat()

	settings = LoadSettings("data/lewdbot.db")
	defer settings.Close()

	client := steam.NewClient(
		settings,
		configuration.Username,
		configuration.Password,
		configuration.Master,
		GenerateReply,
		cleanMessage,
	)
	client.Main()
}
