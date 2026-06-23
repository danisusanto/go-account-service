package event

const UserCreatedEvent = "account.user.created"

type UserCreated struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
}
