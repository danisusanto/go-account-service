package event

const UserEmailChangedEvent = "account.user.email_changed"

type UserEmailChanged struct {
	UserID int `json:"user_id"`
}
