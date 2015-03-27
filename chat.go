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

func learnFileLines(path string, simple bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	text := ""

	s := bufio.NewScanner(bufio.NewReader(f))
	for s.Scan() {
		line := s.Text()
		if !simple { //Learn all lines between empty lines
			if line == "" {
				learnMessage(text)
				text = ""
			} else {
				text += " " + line
			}
		} else { // Learn every line
			learnMessage(line)
		}
	}

	return s.Err()
}

func learnMessage(text string) {
	text = cleanMessage(text)

	if len(text) < 5 ||
		strings.Count(text, " ") < 2 ||
		regex.JustPunctuation.MatchString(text) ||
		regex.LeadingNumbers.MatchString(text) {
		return
	}

	lewdbrain.Learn(text)
}

func GenerateReply(message string) string {
	reply := lewdbrain.Reply(message)
	reply = strings.TrimSpace(reply)

	reply = regex.TrailingPunctuation.ReplaceAllString(reply, "")
	reply = fmt.Sprintf("%s~", reply)

	// TODO: Stop the cancer
	learnMessage(message)

	return reply
}

func init_chat() {
	model := fate.NewModel(fate.Config{})

	lewdbrain = model

	learnFileLines("./data/brain.txt", true)
	learnFileLines("./chatlog.txt", false)
}
