package user

type User struct {
	ListBrowser []string `json:"browsers,"`
	Name        string   `json:"name"`
	Email       string   `json:"email"`
}
