package entities

type User struct {
	ID        int
	Username  string
	Email     string
	Firstname string
	Lastname  string
	AvatarURL string
	Provider  string
	// CreatedAt time.Time
}

type UsersInterface interface {
	GetUser() string
	GetUserInfo(string) User
}
