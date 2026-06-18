package domain

import "time"

type Notification struct {
	ID        uint      `json:"id"`
	UserID    string    `json:"user_id"`
	Message   string    `json:"message"`
	Type      string    `json:"type"` // "EMAIL" atau "SMS"
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}