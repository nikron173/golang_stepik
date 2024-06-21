package main

import (
	"fmt"
	"strings"
)

// сюда писать код
// на сервер грузить только этот файл

var (
	myRoom  *Room
	cook    *Room
	street  *Room
	hallway *Room
)

func initGame() {
	myRoom = NewRoom("комната")
	cook = NewRoom("кухня")
	street = NewRoom("улица")
	hallway = NewRoom("коридор")
	myRoom.AddAvailableRooms(hallway)
	cook.AddAvailableRooms(hallway)
	street.AddAvailableRooms(hallway)
	hallway.AddAvailableRooms(cook, myRoom, street)

	//на столе
	myRoom.AddItemInRoom(
		NewItem("ключи", "стол"),
	)
	myRoom.AddItemInRoom(
		NewItem("конспекты", "стол"),
	)

	//на стуле
	myRoom.AddItemInRoom(
		NewItem("рюкзак", "стул"),
	)

	hallway.AddItemInRoom(
		NewDoor(),
	)

	cook.AddItemInRoom(
		NewItem("чай", "стол"),
	)
	initActionsCook()
	initActionsHallway()
	initActionsMyRoom()
	initActionsStreet()
	// fmt.Printf("Cook: %#v\n", cook)
}

func addPlayer(player *Player) {
	cook.AddPlayer(player)
	player.Room = cook
	// fmt.Printf("Cook: %#v\n", cook)
	// fmt.Printf("Player: %#v\n", player)
}

func initActionsCook() {
	action := make(map[string]func([]string) string)
	action["осмотреться"] = func(in []string) string {
		var builder strings.Builder
		player, _ := cook.GetPlayer(in[0])
		builder.WriteString("ты находишься на кухне, на столе чай, надо ")
		_, ok := player.Items["рюкзак"]
		if !ok {
			builder.WriteString("собрать рюкзак и ")
		}
		builder.WriteString("идти в универ. ")
		builder.WriteString(printAvailableRoom(cook))
		builder.WriteString(printAvailablePlayers(cook, in[len(in)-1]))
		return builder.String()
	}
	action["идти"] = func(in []string) string {
		var builder strings.Builder
		room, ok := cook.GetAvailableRoom(in[0])
		if !ok {
			builder.WriteString(fmt.Sprintf("нет пути в %s", in[0]))
			return builder.String()
		}
		player := cook.RemovePlayer(in[1])
		player.Room = room
		room.AddPlayer(player)
		a := room.Action["пришел"]
		builder.WriteString(a(make([]string, 0)))
		builder.WriteString(printAvailableRoom(room))
		builder.WriteString(printAvailablePlayers(cook, in[len(in)-1]))
		return builder.String()
	}
	action["пришел"] = func(in []string) string {
		return "кухня, ничего интересного. "
	}
	cook.Action = action
}

func initActionsMyRoom() {
	action := make(map[string]func([]string) string)
	action["осмотреться"] = func(in []string) string {
		var builder strings.Builder
		if len(myRoom.ItemsInRoom) == 0 {
			builder.WriteString("пустая комната. ")
			builder.WriteString(printAvailableRoom(myRoom))
			return builder.String()
		}
		itemsTable := myRoom.GetItemInRoomWhereLocation("стол")
		itemsStul := myRoom.GetItemInRoomWhereLocation("стул")
		if len(itemsTable) > 0 {
			builder.WriteString("на столе: ")
			for ind, it := range itemsTable {
				if ind == len(itemsTable)-1 {
					if len(itemsStul) == 0 {
						builder.WriteString(it.Name)
						builder.WriteString(". ")
					} else {
						builder.WriteString(it.Name)
						builder.WriteString(", ")
					}
					continue
				}
				builder.WriteString(it.Name)
				builder.WriteString(", ")
			}
		}
		if len(itemsStul) == 1 {
			builder.WriteString("на стуле - рюкзак. ")
		}
		builder.WriteString(printAvailableRoom(myRoom))
		builder.WriteString(printAvailablePlayers(myRoom, in[len(in)-1]))
		return builder.String()
	}
	action["идти"] = func(in []string) string {
		var builder strings.Builder
		room, ok := myRoom.GetAvailableRoom(in[0])
		if !ok {
			builder.WriteString(fmt.Sprintf("нет пути в %s", in[0]))
			return builder.String()
		}
		player := myRoom.RemovePlayer(in[1])
		player.Room = room
		room.AddPlayer(player)
		a := room.Action["пришел"]
		builder.WriteString(a(make([]string, 0)))
		builder.WriteString(printAvailableRoom(room))
		builder.WriteString(printAvailablePlayers(myRoom, in[len(in)-1]))
		return builder.String()
	}
	action["пришел"] = func(in []string) string {
		return "ты в своей комнате. "
	}
	action["взять"] = func(in []string) string {
		item, ok := myRoom.GetItemInRoom(in[0])
		if !ok {
			return "нет такого"
		}
		if !item.IsMove {
			return "Предмет не может быть взят"
		}
		player, _ := myRoom.GetPlayer(in[1])
		if player.Items["рюкзак"] == nil {
			return "некуда класть"
		}
		player.Items[item.Name] = item
		myRoom.RemoveItemInRoom(item.Name)
		return fmt.Sprintf("предмет добавлен в инвентарь: %s", item.Name)
	}
	action["одеть"] = func(in []string) string {
		item, ok := myRoom.GetItemInRoom(in[0])
		if !ok {
			return "нет такого"
		}
		if !item.IsMove {
			return "Предмет не может быть взят"
		}
		player, _ := myRoom.GetPlayer(in[1])
		player.Items[item.Name] = item
		myRoom.RemoveItemInRoom(item.Name)
		return fmt.Sprintf("вы одели: %s", item.Name)
	}
	myRoom.Action = action
}

func initActionsHallway() {
	action := make(map[string]func([]string) string)
	action["осмотреться"] = func(in []string) string {
		var builder strings.Builder
		builder.WriteString("ничего интересного. ")
		builder.WriteString(printAvailableRoom(hallway))
		builder.WriteString(printAvailablePlayers(hallway, in[len(in)-1]))
		return builder.String()
	}
	action["идти"] = func(in []string) string {
		var builder strings.Builder
		room, ok := hallway.GetAvailableRoom(in[0])
		if !ok {
			builder.WriteString(fmt.Sprintf("нет пути в %s", in[0]))
			return builder.String()
		}
		door, _ := hallway.GetItemInRoom("дверь")
		if room.Name == "улица" && !door.IsOpen {
			return "дверь закрыта"
		}
		player := hallway.RemovePlayer(in[1])
		player.Room = room
		room.AddPlayer(player)
		a := room.Action["пришел"]
		builder.WriteString(a(make([]string, 0)))
		if room.Name == "улица" {
			builder.WriteString("можно пройти - домой")
			return builder.String()
		}
		builder.WriteString(printAvailableRoom(room))
		builder.WriteString(printAvailablePlayers(hallway, in[len(in)-1]))
		return builder.String()
	}
	action["пришел"] = func(in []string) string {
		return "ничего интересного. "
	}
	action["применить"] = func(in []string) string {
		player, _ := hallway.GetPlayer(in[2])
		item, ok := player.Items[in[0]]
		if !ok {
			return fmt.Sprintf("нет предмета в инвентаре - %s", in[0])
		}
		itemInRoom, ok := hallway.GetItemInRoom(in[1])
		if !ok {
			return "не к чему применить"
		}
		if item.Name == "ключи" && itemInRoom.Name == "дверь" {
			itemInRoom.IsOpen = !itemInRoom.IsOpen
			if itemInRoom.IsOpen {
				return "дверь открыта"
			} else {
				return "дверь закрыта"
			}
		}
		return ""
	}
	hallway.Action = action
}

func initActionsStreet() {
	action := make(map[string]func([]string) string)
	action["пришел"] = func(in []string) string {
		return "на улице весна. "
	}
	street.Action = action
}

func printAvailableRoom(room *Room) string {
	var builder strings.Builder
	builder.WriteString("можно пройти - ")
	room.mu.RLock()
	defer room.mu.RUnlock()
	for ind, r := range room.AvailableRooms {
		if ind == len(room.AvailableRooms)-1 {
			builder.WriteString(r.Name)
			continue
		}
		builder.WriteString(r.Name)
		builder.WriteString(", ")
	}
	return builder.String()
}

func printAvailablePlayers(room *Room, currentPlayerName string) string {
	var builder strings.Builder
	room.mu.RLock()
	defer room.mu.RUnlock()
	if len(room.Players) < 2 {
		return ""
	}
	builder.WriteString(". Кроме вас тут ещё ")
	count := 0
	for _, player := range room.Players {
		count++
		if currentPlayerName == player.Name {
			continue
		}
		if len(room.Players) == count {
			builder.WriteString(player.Name)
			continue
		}
		builder.WriteString(player.Name)

	}
	return builder.String()
}
