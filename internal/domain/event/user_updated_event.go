package event

const UserUpdatedEvent = "account.user.updated"

type UserUpdated struct {
	UserID int `json:"user_id"`
}
