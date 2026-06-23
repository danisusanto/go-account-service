package email

const ResetPasswordSuccessSubject = "Your Password Has Been Reset"
const ResetPasswordSuccessTemplate = "reset_password_success.html"

type ResetPasswordSuccess struct {
	Name     string `json:"name"`
	LoginURL string `json:"login_url"`
}
