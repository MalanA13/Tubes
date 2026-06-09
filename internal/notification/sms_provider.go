package notification

import "fmt"

// SendSMS simulates sending an SMS notification.
// In production, this would integrate with Twilio, Vonage, etc.
func SendSMS(to, message string) error {
	fmt.Printf(">>> SMS TO %s | Message: %s\n", to, message)
	return nil
}