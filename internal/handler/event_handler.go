package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/tubes-cc/logistics/internal/notification"
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
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req notifyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(notifyResponse{
				Status:    "error",
				Message:   "invalid request body",
				Timestamp: time.Now().Format(time.RFC3339),
			})
			return
		}

		// Validate required fields
		if req.UserID == "" || req.Message == "" || req.Type == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(notifyResponse{
				Status:    "error",
				Message:   "user_id, message, and type are required",
				Timestamp: time.Now().Format(time.RFC3339),
			})
			return
		}

		// Call the real service — sends notification and persists the log
		if err := s.CreateNotification(req.UserID, req.Message, req.Type); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(notifyResponse{
				Status:    "error",
				Message:   err.Error(),
				Timestamp: time.Now().Format(time.RFC3339),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(notifyResponse{
			Status:    "sent",
			Message:   "notification sent successfully",
			Timestamp: time.Now().Format(time.RFC3339),
		})
	}
}