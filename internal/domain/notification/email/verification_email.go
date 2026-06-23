package email

const EmailVerificationSubject = "Verify Your Email Address"
const EmailVerificationTemplate = "email_verification.html"

type EmailVerification struct {
	Name            string `json:"name"`
	VerificationURL string `json:"verification_url"`
	ExpiredIn       string `json:"expired_in"`
}
