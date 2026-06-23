package email

const ResetPasswordSubject = "Reset Your Password"
const ResetPasswordTemplate = "reset_password.html"

type ResetPassword struct {
	Name      string `json:"name"`
	Token     string `json:"token"`
	ExpiredIn string `json:"expired_in"`
}
