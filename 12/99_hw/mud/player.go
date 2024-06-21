package main

import (
	"strings"
)

type Player struct {
	Name   string
	output chan string
	Room   *Room
	Items  map[string]*Item
}

func NewPlayer(name string) *Player {
	return &Player{
		Name:   name,
		output: make(chan string),
		Room:   nil,
		Items:  make(map[string]*Item),
	}
}

func (p *Player) GetOutput() chan string {
	return p.output
}

func (p *Player) HandleInput(cmd string) {
	parsedCmd := strings.Split(cmd, " ")
	switch parsedCmd[0] {
	case "сказать_игроку":
		{
			p.Room.SendMsgToPlayer(p.Name, parsedCmd[1], strings.Join(parsedCmd[2:], " "))
		}
	case "сказать":
		{
			p.Room.SendMsgBroadcast(p.Name, strings.Join(parsedCmd[1:], " "))
		}
	default:
		{
			action, ok := p.Room.Action[parsedCmd[0]]
			if !ok {
				go func() {
					p.output <- "неизвестная команда"
				}()
				return
			}
			parsedCmd = append(parsedCmd, p.Name)
			resp := action(parsedCmd[1:])
			go func() {
				p.output <- resp
			}()
		}
	}
}
