package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/Philipp15b/go-steam"
	"github.com/Philipp15b/go-steam/internal/steamlang"
	"github.com/Philipp15b/go-steam/socialcache"
	"github.com/Philipp15b/go-steam/steamid"
	cobe "github.com/pteichman/go.cobe"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"github.com/softashell/lewdbot/settings"
	"database/sql"
)

type Configuration struct {
	Username string
	Password string
}

var configuration Configuration
var database *sql.DB

var lewdbrain *cobe.Cobe2Brain
var StrangerList = socialcache.NewFriendsList()

func learnFileLines(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	s := bufio.NewScanner(bufio.NewReader(f))
	for s.Scan() {
		text := CleanMessage(s.Text())
		log.Print(text)
		//lewdbrain.Learn(text)
		/*
			sentences := regexp.MustCompile(`(\.+?[\n\.!?]+)`).Split(text, -1)
			for index, sentence := range sentences {
				sentence = strings.TrimSpace(sentence)
				if len(sentence) < 2 {
					continue
				}
				log.Printf("[%d] %s", index, sentence)
			}
		*/
	}

	return nil
}

func IsRussian(message string) bool {
	if regexp.MustCompile(`\p{Cyrillic}`).MatchString(message) {
		return true
	}

	return false
}

func IsChatRoom(steamid steamid.SteamId) bool {
	if steamid.ToString() != "0" {
		return true
	}

	return false
}

func CleanMessage(message string) string {
	message = regexp.MustCompile(`(https?:\/\/[^\s]+)`).ReplaceAllString(message, "")
	message = regexp.MustCompile(`((:|ː)\w+(:|ː))`).ReplaceAllString(message, "")
	message = regexp.MustCompile(`[:"']`).ReplaceAllString(message, "")

	// GET OUT OF HERE STALKER
	message = regexp.MustCompile(`\p{Cyrillic}`).ReplaceAllString(message, "")

	// Repeated whitespace
	message = regexp.MustCompile(`\s+/`).ReplaceAllString(message, "")

	return strings.TrimSpace(message)
}

func GenerateReply(client *steam.Client, steamid steamid.SteamId, message string) string {
	reply := lewdbrain.Reply(message)
	reply = strings.TrimSpace(reply)

	reply = strings.Replace(reply, "lewdbot", GetName(client, steamid), -1)
	reply = regexp.MustCompile(`[\.—-]+$`).ReplaceAllString(reply, "~")

	// TODO: Stop the cancer
	lewdbrain.Learn(message)

	return reply
}

func ReplyToMessage(client *steam.Client, e *steam.ChatMsgEvent) {
	if !e.IsMessage() {
		return
	}

	if IsRussian(e.Message) && !IsChatRoom(e.ChatRoomId) { // Get out of here stalker
		client.Social.SendMessage(e.ChatterId, steamlang.EChatEntryType_ChatMsg, "Иди нахуй")
		return
	}

	message := CleanMessage(e.Message)

	if len(regexp.MustCompile(`[^\p{L} ]`).ReplaceAllString(message, "")) < 3 { // Not enough actual text to bother replying
		if !IsChatRoom(e.ChatRoomId) {
			client.Social.SendMessage(e.ChatterId, steamlang.EChatEntryType_ChatMsg, "Are you retarded?~")
		}
		return
	}

	reply := GenerateReply(client, e.ChatterId, message)

	LogMessage(client, e.ChatterId, message, reply)

	if IsChatRoom(e.ChatRoomId) { // Group chat
		client.Social.SendMessage(e.ChatRoomId, steamlang.EChatEntryType_ChatMsg, reply)
	} else { // Private message
		client.Social.SendMessage(e.ChatterId, steamlang.EChatEntryType_ChatMsg, reply)
	}
}

func LogMessage(client *steam.Client, chatterid steamid.SteamId, message string, reply string) {
	name := GetName(client, chatterid)
	filename := fmt.Sprintf("%d", chatterid.ToUint64())
	text := fmt.Sprintf("%s: %s\nlewdbot: %s\n", name, message, reply)

	log.Print(text)

	f, err := os.OpenFile("chatlog.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err = f.WriteString(fmt.Sprintf("%s\n", message)); err != nil {
		log.Print(err)
	}

	f, err = os.OpenFile(fmt.Sprintf("./logs/%s.txt", filename), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err = f.WriteString(text); err != nil {
		log.Print(err)
	}
}

func FriendState(client *steam.Client, e *steam.FriendStateEvent) {
	switch e.Relationship {
	case steamlang.EFriendRelationship_None:
		log.Printf("%s removed me from friends list", e.SteamId)
	case steamlang.EFriendRelationship_PendingInvitee:
		log.Printf("%s added me to friends list", GetName(client, e.SteamId))
		client.Social.AddFriend(e.SteamId)
	case steamlang.EFriendRelationship_Friend:
		log.Printf("%s is now a friend", GetName(client, e.SteamId))
		StrangerList.Remove(e.SteamId)
	}
}

// Adds friends who added bot while it was offline
func AddFriends(client *steam.Client, e *steam.FriendsListEvent) {
	for id, friend := range client.Social.Friends.GetCopy() {
		switch friend.Relationship {
		case steamlang.EFriendRelationship_RequestInitiator:
			log.Printf("%s %s still hasn't accepted invite, consider removing", GetName(client, id), id)
		case steamlang.EFriendRelationship_PendingInvitee:
			log.Printf("%s %s added me to friends list", GetName(client, id), id)
			client.Social.AddFriend(id)
		}
	}
}

func ChatInviteEvent(client *steam.Client, e *steam.ChatInviteEvent) {
	if e.ChatRoomType != steamlang.EChatRoomType_Lobby {
		log.Printf("Invited to %s (%s) by %s %s", e.ChatRoomName, e.ChatRoomId, GetName(client, e.PatronId), e.PatronId)
		client.Social.SendMessage(e.PatronId, steamlang.EChatEntryType_ChatMsg, "On my way~ I hope you will not keep me in your basement forever~")
		client.Social.JoinChat(e.ChatRoomId)
	}
}

func ChatEnterEvent(client *steam.Client, e *steam.ChatEnterEvent) {
	if e.EnterResponse == steamlang.EChatRoomEnterResponse_Success {
		log.Printf("Joined %s (%s)", e.Name, e.ChatRoomId)
	} else {
		log.Printf("Failed to join %s! Respone: %s", e.ChatRoomId, e.EnterResponse)
	}
}

func ChatMemberInfo(client *steam.Client, e *steam.ChatMemberInfoEvent) {
	if e.Type == steamlang.EChatInfoType_StateChange {
		if e.StateChangeInfo.ChatterActedOn == client.SteamId() {
			switch e.StateChangeInfo.StateChange {
			case steamlang.EChatMemberStateChange_Kicked:
				log.Printf("Kicked from %s by %s", e.ChatRoomId, GetName(client, e.StateChangeInfo.ChatterActedBy))
			case steamlang.EChatMemberStateChange_Banned:
				log.Printf("Kicked and banned from %s by %s", e.ChatRoomId, GetName(client, e.StateChangeInfo.ChatterActedBy))
			}
		}
	}
}

func GetName(client *steam.Client, friendid steamid.SteamId) string {
	nerd, err := client.Social.Friends.ById(friendid)
	if err == nil {
		return nerd.Name
	}

	nerd, err = StrangerList.ById(friendid)
	if err == nil {
		return nerd.Name
	}

	return "Nerdgin"
}

func PersonaStateEvent(client *steam.Client, e *steam.PersonaStateEvent) {

	if e.FriendId == client.SteamId() {
		return // Updating own status
	}

	_, err := client.Social.Friends.ById(e.FriendId)
	if err == nil {
		return // Is a friend already, no need to update manually
	}

	if e.State == steamlang.EPersonaState_Offline {
		return // Most likely a group update
	}

	StrangerList.Add(
		socialcache.Friend{e.FriendId, e.Name, e.Avatar, steamlang.EFriendRelationship_None,
			e.State, e.StateFlags, e.GameAppId, e.GameId, e.GameName})

	if e.StatusFlags&steamlang.EClientPersonaStateFlag_PlayerName != 0 {
		StrangerList.SetName(e.FriendId, e.Name)
	}

}

func main() {

	os.Mkdir("./data", 0777)
	os.Mkdir("./logs", 0777)

	cobebrain, err := cobe.OpenCobe2Brain("./data/lewdbot.brain")
	if err != nil {
		log.Fatalf("Opening brain file: %s", err)
	}
	defer cobebrain.Close()

	lewdbrain = cobebrain

	//learnFileLines("./data/brain.txt")
	//learnFileLines("chatlog.txt")

	file, _ := os.Open("./config.json")
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&configuration); err != nil {
		log.Fatal(err)
	}

	database = settings.Load()
	defer database.Close()

	myLoginInfo := new(steam.LogOnDetails)
	myLoginInfo.Username = configuration.Username
	myLoginInfo.Password = configuration.Password

	client := steam.NewClient()
	client.Connect()
	defer client.Disconnect()

	for event := range client.Events() {
		switch e := event.(type) {
		case *steam.ConnectedEvent:
			log.Print("Connecting")
			client.Auth.LogOn(myLoginInfo)
		case *steam.MachineAuthUpdateEvent:
			ioutil.WriteFile("sentry", e.Hash, 0666)
		case *steam.LoggedOnEvent:
			log.Print("Logged on")
			client.Social.SetPersonaState(steamlang.EPersonaState_Online)
		case steam.FatalErrorEvent:
			log.Print(e)
		case *steam.ChatMsgEvent:
			go ReplyToMessage(client, e)
		case *steam.FriendStateEvent:
			go FriendState(client, e)
		case *steam.FriendsListEvent:
			go AddFriends(client, e)
		case *steam.ChatInviteEvent:
			go ChatInviteEvent(client, e)
		case *steam.ChatEnterEvent:
			go ChatEnterEvent(client, e)
		case *steam.ChatMemberInfoEvent:
			go ChatMemberInfo(client, e)
		case *steam.FriendAddedEvent:
			client.Social.SendMessage(e.SteamId, steamlang.EChatEntryType_ChatMsg, "Looking forward to working with you~ fu fu fu~")
		case *steam.PersonaStateEvent:
			go PersonaStateEvent(client, e)
		case error:
			log.Print(e)
		}
	}
}
