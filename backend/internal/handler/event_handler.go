package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/tubes-cc/logistics/internal/notification"
	"github.com/tubes-cc/logistics/internal/response"
)

// notifyRequest represents the JSON body for POST /notify.
type notifyRequest struct {
	UserID  string `json:"user_id"`
	Message string `json:"message"`
	Type    string `json:"type"` // "EMAIL" or "SMS"
}

// notifyResponse represents the JSON response after notification is sent.
type notifyResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// HandleSendNotificationHTTP creates an HTTP handler that decodes the request,
// calls the NotificationService to send + persist the notification,
// and returns a proper JSON response.
func HandleSendNotificationHTTP(s notification.NotificationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.MethodNotAllowed(w, "method not allowed")
			return
		}

		var req notifyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, "invalid request body")
			return
		}

		// Validate required fields
		if req.UserID == "" || req.Message == "" || req.Type == "" {
			response.BadRequest(w, "user_id, message, and type are required")
			return
		}

		// Call the real service — sends notification and persists the log
		if err := s.CreateNotification(req.UserID, req.Message, req.Type); err != nil {
			response.InternalServerError(w, err.Error())
			return
		}

		response.Created(w, notifyResponse{
			Status:    "sent",
			Message:   "notification sent successfully",
			Timestamp: time.Now().Format(time.RFC3339),
		})
	}
}