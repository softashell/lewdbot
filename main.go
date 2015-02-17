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
	"github.com/softashell/lewdbot/commands"
	"github.com/softashell/lewdbot/regex"
	. "github.com/softashell/lewdbot/settings"
	"io/ioutil"
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
var settings Settings

var lewdbrain *cobe.Cobe2Brain
var StrangerList = socialcache.NewFriendsList()

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

func GenerateReply(client *steam.Client, steamid steamid.SteamId, message string) string {
	reply := lewdbrain.Reply(message)
	reply = strings.TrimSpace(reply)

	reply = strings.Replace(reply, "lewdbot", GetName(client, steamid), 1)
	reply = regex.TrailingPunctuation.ReplaceAllString(reply, "")
	reply = fmt.Sprintf("%s~", reply)

	// TODO: Stop the cancer
	lewdbrain.Learn(message)

	return reply
}

func ReplyToMessage(client *steam.Client, e *steam.ChatMsgEvent) {
	if !e.IsMessage() {
		return
	}

	if isMaster(e.ChatterId) {
		if master, replies := commands.Handle(client, e.Message, settings); master == true {
			for _, reply := range replies {
				client.Social.SendMessage(e.ChatterId, steamlang.EChatEntryType_ChatMsg, reply)
			}
			if len(replies) == 0 {
				client.Social.SendMessage(e.ChatterId, steamlang.EChatEntryType_ChatMsg, "I got nothing!")
			}
			return
		}
	}

	message := e.Message

	if isChatRoom(e.ChatRoomId) {
		if strings.HasPrefix(strings.ToLower(e.Message), "lewdbot, ") {
			switch {
			case strings.HasSuffix(e.Message, "don't speak unless spoken to."):
				settings.SetGroupQuiet(e.ChatRoomId, true)
				client.Social.SendMessage(e.ChatRoomId, steamlang.EChatEntryType_ChatMsg, "Got it!")
				return
			case strings.HasSuffix(e.Message, "you may speak freely."):
				settings.SetGroupQuiet(e.ChatRoomId, false)
				client.Social.SendMessage(e.ChatRoomId, steamlang.EChatEntryType_ChatMsg, "Got it!")
				return
			case strings.HasSuffix(e.Message, "you can come here any time you'd like."):
				settings.SetGroupAutojoin(e.ChatRoomId, true)
				client.Social.SendMessage(e.ChatRoomId, steamlang.EChatEntryType_ChatMsg, "Got it!")
				return
			case strings.HasSuffix(e.Message, "stop coming here."):
				settings.SetGroupAutojoin(e.ChatRoomId, false)
				client.Social.SendMessage(e.ChatRoomId, steamlang.EChatEntryType_ChatMsg, "Got it!")
				return
			default:
				message = message[9:]
			}
		} else {
			if settings.IsGroupQuiet(e.ChatRoomId) {
				// todo: logmessage here, without a reply
				return
			}
		}
	}

	if isRussian(e.Message) { // Should be called before cleanMessage since it replaces russian
		if !isChatRoom(e.ChatRoomId) { // Get out of here stalker
			client.Social.SendMessage(e.ChatterId, steamlang.EChatEntryType_ChatMsg, "Иди нахуй")
		}
		return
	}

	message = cleanMessage(message)
	var destination steamid.SteamId
	if isChatRoom(e.ChatRoomId) {
		destination = e.ChatRoomId
	} else {
		destination = e.ChatterId
	}

	if len(regex.NotActualText.ReplaceAllString(message, "")) < 3 { // Not enough actual text to bother replying
		if !isChatRoom(e.ChatRoomId) {
			client.Social.SendMessage(e.ChatterId, steamlang.EChatEntryType_ChatMsg, "Are you retarded?~")
		}
		return
	} else if regex.Greentext.MatchString(message) {
		client.Social.SendMessage(destination, steamlang.EChatEntryType_ChatMsg, "Who are you quoting?~")
		return
	} else if regex.JustPunctuation.MatchString(message) {
		return
	}

	reply := GenerateReply(client, e.ChatterId, message)

	LogMessage(client, destination, e.ChatterId, message, reply)
	client.Social.SendMessage(destination, steamlang.EChatEntryType_ChatMsg, reply)
}

func LogMessage(client *steam.Client, id steamid.SteamId, chatter steamid.SteamId, message string, reply string) {
	name := "Nerdgin"
	if !isChatRoom(id) {
		name = GetName(client, id)
	} else {
		name = GetName(client, chatter)
	}

	filename := fmt.Sprintf("%d", id.ToUint64())
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
		log.Printf("%s removed me from friends list", steamLink(e.SteamId))
	case steamlang.EFriendRelationship_PendingInvitee:
		log.Printf("%s added me to friends list", steamLink(e.SteamId))
		client.Social.AddFriend(e.SteamId)
	case steamlang.EFriendRelationship_Friend:
		log.Printf("%s (%s) is now a friend", GetName(client, e.SteamId), steamLink(e.SteamId))
		StrangerList.Remove(e.SteamId)
	}
}

// Adds friends who added bot while it was offline
func AddFriends(client *steam.Client, e *steam.FriendsListEvent) {
	for id, friend := range client.Social.Friends.GetCopy() {
		switch friend.Relationship {
		case steamlang.EFriendRelationship_RequestInitiator:
			log.Printf("%s still hasn't accepted invite, consider removing", steamLink(id))
		case steamlang.EFriendRelationship_PendingInvitee:
			log.Printf("%s added me to friends list while I was offline", steamLink(id))
			client.Social.AddFriend(id)
		}
	}
}

func ChatInviteEvent(client *steam.Client, e *steam.ChatInviteEvent) {
	if e.ChatRoomType != steamlang.EChatRoomType_Lobby {
		log.Printf("Invited to %s (%s) by %s %d", e.ChatRoomName, e.ChatRoomId, GetName(client, e.PatronId), e.PatronId.ToUint64())

		if !settings.IsGroupBlacklisted(e.ChatRoomId) {
			client.Social.SendMessage(e.PatronId, steamlang.EChatEntryType_ChatMsg, "On my way~ I hope you will not keep me in your basement forever~")
			client.Social.JoinChat(e.ChatRoomId)
		} else {
			log.Print("group is blacklisted ")
			client.Social.SendMessage(e.PatronId, steamlang.EChatEntryType_ChatMsg, "Only disgusting nerds go there~")
		}
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
			case steamlang.EChatMemberStateChange_Left:
				log.Printf("Left room http://steamcommunity.com/gid/%d", e.ChatRoomId)
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

func AutojoinGroups(client *steam.Client) {
	autojoin := settings.ListGroupAutojoin()
	for _, group := range autojoin {
		client.Social.JoinChat(group)
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

	file, _ := os.Open("./config.json")
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&configuration); err != nil {
		log.Fatal(err)
	}

	settings = LoadSettings("data/lewdbot.db")
	defer settings.Close()

	myLoginInfo := new(steam.LogOnDetails)
	myLoginInfo.Username = configuration.Username
	myLoginInfo.Password = configuration.Password

	client := steam.NewClient()
	client.Connect()
	//defer client.Disconnect()

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
			go AutojoinGroups(client)
		case *steam.DisconnectedEvent:
			log.Print("DisconnectedEvent: ", e)
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
		case steam.FatalErrorEvent:
			log.Print("FatalErrorEvent: ", e)
		case error:
			log.Print("error: ", e)
			if client.Connected() {
				log.Print("not attempting to reconnect")
			} else {
				log.Print("attempting to reconnect")
				client.Connect()
			}
		}
	}
}
