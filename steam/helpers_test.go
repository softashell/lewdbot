package steam

import (
	"strings"
	"testing"
)

func TestIsRussian(t *testing.T) {
	if isRussian("Hello") != false {
		t.Error("False positive")
	}
	if isRussian("Иди нахуй") != true {
		t.Error("Didn't detect")
	}
}

func TestIsChatRoom(t *testing.T) {
	c := Client{}
	if c.isChatRoom(76561197983301654) != false {
		t.Error("False positive") //broken? v(´・ω・｀)v
	}
	if c.isChatRoom(103582791435317007) != true {
		t.Error("Didn't detect")
	}
}

func TestSteamLink(t *testing.T) {
	c := Client{}
	if !strings.Contains(c.link(76561197983301654), "profiles") {
		t.Error("Profiles link didn't work")
	}
	if !strings.Contains(c.link(103582791435317007), "gid") {
		t.Error("Group link didn't work")
	}
}

// todo: add tests for cleanMessage
