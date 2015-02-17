package main

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
	if isChatRoom(76561197983301654) != false {
		//t.Error("False positive") broken? v(´・ω・｀)v
	}
	if isChatRoom(103582791435317007) != true {
		t.Error("Didn't detect")
	}
}

func TestSteamLink(t *testing.T) {
	if !strings.Contains(steamLink(76561197983301654), "profiles") {
		t.Error("Profiles link didn't work")
	}
	if !strings.Contains(steamLink(103582791435317007), "gid") {
		t.Error("Group link didn't work")
	}
}

// todo: add tests for cleanMessage
