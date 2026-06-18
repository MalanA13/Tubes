package notification

import models "github.com/tubes-cc/logistics/domain"

type NotificationRepository interface {
	Send(notif *models.Notification) error
	SaveLog(notif *models.Notification) error
}