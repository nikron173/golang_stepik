package models

type Task struct {
	ID       int
	Author   *User
	Name     string
	Executor *User
}
