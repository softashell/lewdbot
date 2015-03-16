package main

import (
	"bufio"
	"fmt"
	"github.com/pteichman/fate"
	"github.com/softashell/lewdbot/regex"
	"os"
	"strings"
)

var lewdbrain *fate.Model

func learnFileLines(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	s := bufio.NewScanner(bufio.NewReader(f))
	for s.Scan() {
		text := cleanMessage(s.Text())
		if len(text) < 5 {
			continue
		}
		lewdbrain.Learn(text)
	}

	return s.Err()
}

func GenerateReply(message string) string {
	reply := lewdbrain.Reply(message)
	reply = strings.TrimSpace(reply)

	reply = regex.TrailingPunctuation.ReplaceAllString(reply, "")
	reply = fmt.Sprintf("%s~", reply)

	// TODO: Stop the cancer
	lewdbrain.Learn(message)

	return reply
}

func init_chat() {
	model := fate.NewModel(fate.Config{})

	lewdbrain = model

	learnFileLines("./data/brain.txt")
	learnFileLines("./chatlog.txt")
}
