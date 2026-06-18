package notification

import "fmt"

// SendEmail simulates sending an email notification.
// In production, this would integrate with SendGrid, SES, etc.
func SendEmail(to, subject, body string) error {
	fmt.Printf(">>> EMAIL TO %s | Subject: %s | Body: %s\n", to, subject, body)
	return nil
}