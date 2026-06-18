package notification

import (
	"errors"
	models "github.com/tubes-cc/logistics/domain"
	
)

type NotificationService struct {
	repo NotificationRepository
}

func NewNotificationService(r NotificationRepository) *NotificationService {
	return &NotificationService{repo: r}
}

func (s *NotificationService) CreateNotification(userID, msg, msgType string) error {
	if msg == "" {
		return errors.New("pesan tidak boleh kosong")
	}

	notif := &models.Notification{
		UserID:  userID,
		Message: msg,
		Type:    msgType,
		Status:  "SENT",
	}

	// Memanggil provider (lewat repository)
	err := s.repo.Send(notif)
	if err != nil {
		return err
	}

	return s.repo.SaveLog(notif)
}