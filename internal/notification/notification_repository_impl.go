package notification

import (
	"fmt"
	models "github.com/tubes-cc/logistics/domain"
	"gorm.io/gorm"
)

type notificationRepositoryImpl struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepositoryImpl{db: db}
}

func (r *notificationRepositoryImpl) Send(notif *models.Notification) error {
	// Dalam realita, di sini kamu memanggil API Email/SMS (misal: Twilio atau SendGrid)
	// Untuk kebutuhan tugas fungsional, kita simulasikan dengan print ke konsol
	fmt.Printf(">>> MENGIRIM %s KE %s: %s\n", notif.Type, notif.UserID, notif.Message)
	return nil
}

func (r *notificationRepositoryImpl) SaveLog(notif *models.Notification) error {
	// Menyimpan history notifikasi ke DB
	return r.db.Create(notif).Error
}