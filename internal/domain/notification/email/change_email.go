package email

const ChangeEmailSubject = "Verify Your New Email Address"
const ChangeEmailTemplate = "change_email.html"

type ChangeEmail struct {
	Name      string `json:"name"`
	Token     string `json:"token"`
	ExpiredIn string `json:"expired_in"`
}
