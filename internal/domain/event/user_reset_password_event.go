package event

const UserResetPasswordEvent = "account.user.reset_password"

type UserResetPassword struct {
	UserID int `json:"user_id"`
}
