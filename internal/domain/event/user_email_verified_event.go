package event

const UserEmailVerifiedEvent = "account.user.email_verified"

type UserEmailVerified struct {
	UserID int `json:"user_id"`
}
