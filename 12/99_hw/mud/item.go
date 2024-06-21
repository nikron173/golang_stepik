package main

type Item struct {
	Name     string
	Location string
	IsMove   bool
	IsOpen   bool
}

func NewItem(name, location string) *Item {
	return &Item{
		Name:     name,
		Location: location,
		IsMove:   true,
		IsOpen:   false,
	}
}

func NewDoor() *Item {
	return &Item{
		Name:     "дверь",
		Location: "",
		IsMove:   false,
		IsOpen:   false,
	}
}
