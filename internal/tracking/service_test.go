package tracking

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tubes-cc/logistics/domain"
	"go.uber.org/mock/gomock"
)

func TestTrackingService_AddTrackingEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockTrackingRepository(ctrl)
	service := NewTrackingService(mockRepo)

	t.Run("Successfully add tracking event", func(t *testing.T) {
		resiID := "AWB123456"
		status := domain.StatusInHub
		location := "Hub Jakarta Pusat"
		note := "Package sorted and ready"

		// Mock save event expects any domain.TrackingEvent because time and uuid are dynamic
		mockRepo.EXPECT().SaveEvent(gomock.Any()).DoAndReturn(func(event domain.TrackingEvent) error {
			assert.Equal(t, resiID, event.ResiID)
			assert.Equal(t, status, event.Status)
			assert.Equal(t, location, event.Location)
			assert.Equal(t, note, event.Note)
			assert.NotEmpty(t, event.ID)
			return nil
		})
		mockRepo.EXPECT().UpdateShipmentStatus(resiID, status).Return(nil)

		err := service.AddTrackingEvent(resiID, status, location, note)
		assert.NoError(t, err)
	})

	t.Run("Empty Resi ID", func(t *testing.T) {
		err := service.AddTrackingEvent("", domain.StatusCreated, "Origin", "Created")
		assert.Error(t, err)
		assert.Equal(t, "resi_id cannot be empty", err.Error())
	})

	t.Run("Save Event Error", func(t *testing.T) {
		mockRepo.EXPECT().SaveEvent(gomock.Any()).Return(errors.New("db save error"))

		err := service.AddTrackingEvent("AWB1", domain.StatusInTransit, "Hub A", "")
		assert.Error(t, err)
		assert.Equal(t, "db save error", err.Error())
	})
	
	t.Run("Update Status Error", func(t *testing.T) {
		mockRepo.EXPECT().SaveEvent(gomock.Any()).Return(nil)
		mockRepo.EXPECT().UpdateShipmentStatus("AWB2", domain.StatusDelivered).Return(errors.New("db update error"))

		err := service.AddTrackingEvent("AWB2", domain.StatusDelivered, "Destination", "Delivered")
		assert.Error(t, err)
		assert.Equal(t, "db update error", err.Error())
	})
}

func TestTrackingService_GetHistory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockTrackingRepository(ctrl)
	service := NewTrackingService(mockRepo)

	t.Run("Successfully get history", func(t *testing.T) {
		resiID := "AWB999"
		mockEvents := []domain.TrackingEvent{
			{ID: "uuid-1", ResiID: resiID, Status: domain.StatusCreated, Location: "Store", Note: "Picked up"},
			{ID: "uuid-2", ResiID: resiID, Status: domain.StatusInHub, Location: "Hub A", Note: "Arrived at Hub A"},
		}

		mockRepo.EXPECT().GetTrackingHistory(resiID).Return(mockEvents, nil)

		events, err := service.GetHistory(resiID)
		assert.NoError(t, err)
		assert.Len(t, events, 2)
		assert.Equal(t, "uuid-1", events[0].ID)
	})

	t.Run("Empty Resi ID", func(t *testing.T) {
		events, err := service.GetHistory("")
		assert.Error(t, err)
		assert.Nil(t, events)
		assert.Equal(t, "resi_id cannot be empty", err.Error())
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo.EXPECT().GetTrackingHistory("AWB000").Return(nil, errors.New("not found"))

		events, err := service.GetHistory("AWB000")
		assert.Error(t, err)
		assert.Nil(t, events)
		assert.Equal(t, "not found", err.Error())
	})
}
