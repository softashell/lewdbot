package main

import (
	"encoding/json"
	. "github.com/softashell/lewdbot/settings"
	"log"
	"os"
)

type Configuration struct {
	Username string
	Password string
	Master   uint64
}

var configuration Configuration
var settings Settings

func main() {
	os.Mkdir("./data", 0777)
	os.Mkdir("./logs", 0777)

	file, _ := os.Open("./config.json")
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&configuration); err != nil {
		log.Fatal(err)
	}

	init_cobe()

	settings = LoadSettings("data/lewdbot.db")
	defer settings.Close()

	main_steam()
}
