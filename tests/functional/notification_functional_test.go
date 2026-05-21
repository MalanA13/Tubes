package functional_test

import (
	"testing"
	"github.com/tubes-cc/logistics/internal/notification"
	"github.com/tubes-cc/logistics/domain"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNotification_Functional(t *testing.T) {
	// 1. Setup Database SQLite khusus notifikasi
	db, _ := gorm.Open(sqlite.Open("notification_test.db"), &gorm.Config{})
	db.AutoMigrate(&domain.Notification{})
	db.Exec("DELETE FROM notifications") // Bersihkan data lama

	// 2. Setup Repository dan Service
	notifRepo := notification.NewNotificationRepository(db)
	notifService := notification.NewNotificationService(notifRepo)

	// 3. Jalankan Aksi: Kirim Notifikasi
	err := notifService.CreateNotification("user-007", "Paketmu sedang dikirim kurir!", "SMS")

	// 4. Verifikasi: Cek apakah data tersimpan di DB
	var savedNotif domain.Notification
	db.First(&savedNotif, "user_id = ?", "user-007")

	// 5. Assert
	assert.NoError(t, err)
	assert.Equal(t, "SMS", savedNotif.Type)
	assert.Equal(t, "Paketmu sedang dikirim kurir!", savedNotif.Message)
}