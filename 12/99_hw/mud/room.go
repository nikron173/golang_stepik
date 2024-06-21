package main

import (
	"errors"
	"fmt"
	"sync"
)

var (
	errPlayerNotFound = errors.New("player not found")
)

type Room struct {
	Name           string
	mu             *sync.RWMutex
	wg             *sync.WaitGroup
	Players        map[string]*Player
	AvailableRooms []*Room
	ItemsInRoom    []*Item
	Action         map[string]func([]string) string
}

func NewRoom(nameRoom string) *Room {
	return &Room{
		mu:             &sync.RWMutex{},
		wg:             &sync.WaitGroup{},
		Players:        map[string]*Player{},
		Name:           nameRoom,
		ItemsInRoom:    make([]*Item, 0),
		AvailableRooms: make([]*Room, 0),
		Action:         make(map[string]func([]string) string),
	}
}

func (r *Room) GetItemInRoomWhereLocation(location string) []*Item {
	res := make([]*Item, 0)
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.ItemsInRoom {
		if item.Location == location {
			res = append(res, item)
		}
	}
	return res
}

func (r *Room) AddItemInRoom(item *Item) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ItemsInRoom = append(r.ItemsInRoom, item)
}

func (r *Room) RemoveItemInRoom(nameItem string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	items := make([]*Item, 0, len(r.ItemsInRoom))
	for _, item := range r.ItemsInRoom {
		if item.Name == nameItem {
			continue
		}
		items = append(items, item)
	}
	r.ItemsInRoom = items
}

func (r *Room) GetItemInRoom(nameItem string) (*Item, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, item := range r.ItemsInRoom {
		if item.Name == nameItem {
			return item, true
		}
	}
	return nil, false
}

func (r *Room) AddAvailableRooms(rooms ...*Room) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.AvailableRooms = append(r.AvailableRooms, rooms...)
}

func (r *Room) AddPlayer(player *Player) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Players[player.Name] = player
}

func (r *Room) RemovePlayer(playerName string) *Player {
	r.mu.Lock()
	defer r.mu.Unlock()
	player, ok := r.Players[playerName]
	if !ok {
		return nil
	}
	delete(r.Players, playerName)
	return player
}

func (r *Room) GetPlayer(playerName string) (*Player, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	player, ok := r.Players[playerName]
	if !ok {
		return nil, false
	}
	return player, true
}

func (r *Room) SendMsgToPlayer(playerNameFrom, playerNameTo, msg string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	playerTo, ok := r.Players[playerNameTo]
	playerFrom := r.Players[playerNameFrom]
	if !ok {
		go func() {
			playerFrom.GetOutput() <- "тут нет такого игрока"
		}()
		return
	}
	if msg == "" {
		go func() {
			playerTo.GetOutput() <- fmt.Sprintf("%s выразительно молчит, смотря на вас", playerNameFrom)
		}()
		return
	}
	go func() {
		playerTo.GetOutput() <- fmt.Sprintf("%s говорит вам: %s", playerNameFrom, msg)
	}()
}

func (r *Room) SendMsgBroadcast(playerNameFrom, msg string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, player := range r.Players {
		r.wg.Add(1)
		go func(p *Player) {
			defer r.wg.Done()
			p.GetOutput() <- fmt.Sprintf("%s говорит: %s", playerNameFrom, msg)
		}(player)
	}
	r.wg.Wait()
}

func (r *Room) GetAvailableRoom(room string) (*Room, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, val := range r.AvailableRooms {
		if val.Name == room {
			return val, true
		}
	}
	return nil, false
}
