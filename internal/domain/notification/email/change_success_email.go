package email

const ChangeEmailSuccessSubject = "Your Email Address Has Been Updated"
const ChangeEmailSuccessTemplate = "change_email_success.html"

type ChangeEmailSuccess struct {
	Name     string `json:"name"`
	LoginURL string `json:"login_url"`
}
