package main

import (
	"bufio"
	"fmt"
	"github.com/pteichman/fate"
	"github.com/softashell/lewdbot/regex"
	"math"
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

func learnMessage(text string) bool {
	text = cleanMessage(text)

	if len(text) < 5 ||
		len(text) > 1000 ||
		strings.Count(text, " ") < 2 ||
		regex.JustPunctuation.MatchString(text) ||
		regex.LeadingNumbers.MatchString(text) ||
		getWordCount(text) < 3 ||
		generateEntropy(text) < 3.0 {
		return false // Text doesn't contain enough information
	}

	lewdbrain.Learn(text)

	return true
}

func generateReply(message string) (string, bool) {
	reply := lewdbrain.Reply(message)
	reply = strings.TrimSpace(reply)

	reply = regex.TrailingPunctuation.ReplaceAllString(reply, "")
	reply = fmt.Sprintf("%s~", reply)

	// TODO: Stop the cancer

	return reply, learnMessage(message)
}

func generateEntropy(s string) (e float64) {
	m := make(map[rune]bool)
	for _, r := range s {
		if m[r] {
			continue
		}
		m[r] = true
		n := strings.Count(s, string(r))
		p := float64(n) / float64(len(s))
		e += p * math.Log(p) / math.Log(2)
	}
	return math.Abs(e)
}

func getWordCount(s string) int {
	strs := strings.Fields(s)
	res := make(map[string]int)

	for _, str := range strs {
		res[strings.ToLower(str)]++
	}

	return len(res)
}

func init_chat() {
	model := fate.NewModel(fate.Config{})

	lewdbrain = model

	learnFileLines("./data/brain.txt", true)
	learnFileLines("./chatlog.txt", false)
}
