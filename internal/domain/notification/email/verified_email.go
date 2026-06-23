package email

const EmailVerifiedSubject = "Email Successfully Verified 🎉"
const EmailVerifiedTemplate = "email_verified.html"

type EmailVerified struct {
	Name     string `json:"name"`
	LoginURL string `json:"login_url"`
}
