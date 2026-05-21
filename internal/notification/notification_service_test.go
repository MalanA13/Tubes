package notification

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCreateNotification_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockNotificationRepository(ctrl)
	notifService := NewNotificationService(mockRepo)

	// Ekspektasi: Fungsi Send dan SaveLog harus dipanggil masing-masing 1x
	mockRepo.EXPECT().Send(gomock.Any()).Return(nil).Times(1)
	mockRepo.EXPECT().SaveLog(gomock.Any()).Return(nil).Times(1)

	err := notifService.CreateNotification("user-123", "Paket sampai!", "EMAIL")

	assert.NoError(t, err)
}