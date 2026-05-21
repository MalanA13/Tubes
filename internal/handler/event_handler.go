package handler

import (
	"net/http"
	"github.com/tubes-cc/logistics/internal/notification"
)

// GANTI fungsi lama kamu dengan struktur HTTP Handler yang menerima Service ini:
func HandleSendNotificationHTTP(s notification.NotificationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Logika HTTP kamu masuk di DALAM sini, bang!
		// Contoh pembacaan body request atau logic pengiriman notifikasi kemarin:
		// _ = s.SendEmail(...) 

		// 2. Kirim respon balik ke browser/postman
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "Notification Service is Running!"}`))
	}
}