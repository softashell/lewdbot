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
)

type Configuration struct {
	Username string
	Password string
}

var configuration Configuration

var lewdbrain *cobe.Cobe2Brain
var StrangerList = socialcache.NewFriendsList()

func learnFileLines(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	s := bufio.NewScanner(bufio.NewReader(f))
	for s.Scan() {
		log.Print(s.Text())
		lewdbrain.Learn(CleanMessage(s.Text()))
	}

	return nil
}

func CleanMessage(message string) string {
	url := regexp.MustCompile(`(https?:\/\/[^\s]+)`)
	emote := regexp.MustCompile(`((:|ː)\w+(:|ː))`)

	message = url.ReplaceAllString(message, "")
	message = emote.ReplaceAllString(message, "")

	return strings.TrimSpace(message)
}

func GenerateReply(message string) string {
	reply := lewdbrain.Reply(message)

	fullstop := regexp.MustCompile(`\.+$`)

	reply = fullstop.ReplaceAllString(reply, "~")

	lewdbrain.Learn(message)

	return reply
}

func ReplyToMessage(client *steam.Client, e *steam.ChatMsgEvent) {
	if !e.IsMessage() {
		return
	}

	message := CleanMessage(e.Message)
	reply := GenerateReply(message)

	LogMessage(client, e.ChatterId, message, reply)

	if e.ChatRoomId.ToString() != "0" { // Group chat
		client.Social.SendMessage(e.ChatRoomId, steamlang.EChatEntryType_ChatMsg, reply)
	} else { // Private message
		client.Social.SendMessage(e.ChatterId, steamlang.EChatEntryType_ChatMsg, reply)
	}
}

func LogMessage(client *steam.Client, chatterid steamid.SteamId, message string, reply string) {
	name := GetName(client, chatterid)
	filename := fmt.Sprintf("%s", chatterid.ToUint64())
	text := fmt.Sprintf("%s: %s\nlewdbot: %s\n", name, message, reply)

	log.Printf(text)

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
		log.Print(e.SteamId, " removed me from friends list")
	case steamlang.EFriendRelationship_PendingInvitee:
		log.Print(GetName(client, e.SteamId), " added me to friends list")
		client.Social.AddFriend(e.SteamId)
	case steamlang.EFriendRelationship_Friend:
		log.Print(GetName(client, e.SteamId), " is now a friend")
		StrangerList.Remove(e.SteamId)
	}
}

// Adds friends who added bot while it was offline
func AddFriends(client *steam.Client, e *steam.FriendsListEvent) {
	for id, friend := range client.Social.Friends.GetCopy() {
		switch friend.Relationship {
		case steamlang.EFriendRelationship_RequestInitiator:
			log.Print(GetName(client, id), " (", id, ") still hasn't accepted invite, consider removing")
		case steamlang.EFriendRelationship_PendingInvitee:
			log.Print(GetName(client, id), " (", id, ") added me to friends list")
			client.Social.AddFriend(id)
		}
	}
}

func ChatInviteEvent(client *steam.Client, e *steam.ChatInviteEvent) {
	if e.ChatRoomType != steamlang.EChatRoomType_Lobby {
		log.Print("Invited to ", e.ChatRoomName, " (", e.ChatRoomId, ") by ", GetName(client, e.PatronId), "(", e.PatronId, ")")
		client.Social.SendMessage(e.PatronId, steamlang.EChatEntryType_ChatMsg, "On my way~ I hope you will not keep me in your basement forever~")
		client.Social.JoinChat(e.ChatRoomId)
	}
}

func ChatEnterEvent(client *steam.Client, e *steam.ChatEnterEvent) {
	if e.EnterResponse == steamlang.EChatRoomEnterResponse_Success {
		log.Print("Joined ", e.Name, " (", e.ChatRoomId, ")")
	} else {
		log.Print("Failed to join ", e.EnterResponse)
	}
}

func ChatMemberInfo(client *steam.Client, e *steam.ChatMemberInfoEvent) {
	if e.Type == steamlang.EChatInfoType_StateChange {
		if e.StateChangeInfo.ChatterActedOn == client.SteamId() {
			switch e.StateChangeInfo.StateChange {
			case steamlang.EChatMemberStateChange_Kicked:
				log.Print("Kicked from ", e.ChatRoomId, " by ", GetName(client, e.StateChangeInfo.ChatterActedBy))
			case steamlang.EChatMemberStateChange_Banned:
				log.Print("Kicked and banned from", e.ChatRoomId, " by ", GetName(client, e.StateChangeInfo.ChatterActedBy))
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

	log.Print("Unknown user:", friendid.ToString())

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

	log.Print("PersonaStateEvent: ", e.Name)

	StrangerList.Add(
		socialcache.Friend{e.FriendId, e.Name, e.Avatar, steamlang.EFriendRelationship_None,
			steamlang.EPersonaState_Online, steamlang.EPersonaStateFlag_HasRichPresence, e.GameAppId, e.GameId, e.GameName})

	if e.StatusFlags&steamlang.EClientPersonaStateFlag_PlayerName != 0 {
		StrangerList.SetName(e.FriendId, e.Name)
	}

}

func main() {

	os.Mkdir("./data", 0777)
	os.Mkdir("./logs", 0777)

	cobebrain, err := cobe.OpenCobe2Brain("./data/lewdbot.brain")
	defer cobebrain.Close()

	if err != nil {
		log.Fatalf("Opening brain file: %s", err)
	}

	lewdbrain = cobebrain

	//learnFileLines("brain.txt")

	file, _ := os.Open("./config.json")
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configuration)
	if err != nil {
		log.Fatal(err)
	}

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
			log.Print("FriendAddedEvent:", e)
			client.Social.SendMessage(e.SteamId, steamlang.EChatEntryType_ChatMsg, "Looking forward to working with you~ fu fu fu~")
		case *steam.PersonaStateEvent:
			go PersonaStateEvent(client, e)
		case error:
			log.Print(e)
		}
	}
}
