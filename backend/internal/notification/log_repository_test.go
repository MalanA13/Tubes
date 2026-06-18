package notification

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSendNotification_Success(t *testing.T) {
	// 1. Setup Mock Controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 2. Inisialisasi Mock Repo dan Service Notification
	mockRepo := NewMockNotificationRepository(ctrl)
	notifService := NewNotificationService(mockRepo)

	// 3. Buat dummy data Notification
	// HAPUS ATAU KOMENTARI BARIS DI BAWAH INI:
	// dummyNotif := &models.Notification{} 

	// 4. Set Ekspektasi Mock
	mockRepo.EXPECT().Send(gomock.Any()).Return(nil).Times(1)
	mockRepo.EXPECT().SaveLog(gomock.Any()).Return(nil).Times(1)

	// 5. Jalankan fungsi service aslinya dengan 3 string dummy
	err := notifService.CreateNotification("tes@gmail.com", "Judul Notifikasi", "Isi pesan notifikasi di sini")

	// 6. Validasi harus tidak ada error
	assert.NoError(t, err)
}