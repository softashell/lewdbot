package steam

import (
	. "github.com/softashell/go-steam/steamid"
	"sync"
)

// map[SteamId]SteamId in lewdbot
type InviteList struct {
	mutex sync.RWMutex
	byId  map[SteamId]SteamId
}

func NewInviteList() *InviteList {
	return &InviteList{byId: make(map[SteamId]SteamId)}
}

func (list *InviteList) Add(group SteamId, friend SteamId) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	_, exists := list.byId[group]
	if !exists { //make sure this doesnt already exist
		list.byId[group] = friend
	}
}

func (list *InviteList) Remove(group SteamId) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	delete(list.byId, group)
}

func (list *InviteList) ById(id SteamId) SteamId {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	if val, ok := list.byId[id]; ok {
		return val
	}
	return 0
}
