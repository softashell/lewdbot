package main

import (
	"bufio"
	"fmt"
	"github.com/Philipp15b/go-steam"
	"github.com/Philipp15b/go-steam/steamid"
	cobe "github.com/pteichman/go.cobe"
	"github.com/softashell/lewdbot/regex"
	"log"
	"os"
	"strings"
)

var lewdbrain *cobe.Cobe2Brain

func learnFileLines(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	s := bufio.NewScanner(bufio.NewReader(f))
	for s.Scan() {
		text := cleanMessage(s.Text())
		if len(text) < 5 {
			continue
		}
		lewdbrain.Learn(text)
	}

	return nil
}

// TODO: remove need for *steam.Client and steamid.SteamId arguments
func GenerateReply(client *steam.Client, steamid steamid.SteamId, message string) string {
	reply := lewdbrain.Reply(message)
	reply = strings.TrimSpace(reply)

	reply = strings.Replace(reply, client.Social.GetPersonaName(), steamName(client, steamid), 1)
	reply = regex.TrailingPunctuation.ReplaceAllString(reply, "")
	reply = fmt.Sprintf("%s~", reply)

	// TODO: Stop the cancer
	lewdbrain.Learn(message)

	return reply
}

func init_cobe() {
	cobebrain, err := cobe.OpenCobe2Brain("./data/lewdbot.brain")
	if err != nil {
		log.Fatalf("Opening brain file: %s", err)
	}

	lewdbrain = cobebrain

	//learnFileLines("./data/brain.txt")
}
