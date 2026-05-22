package functional

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tubes-cc/logistics/domain"
	"github.com/tubes-cc/logistics/internal/tracking"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTrackingFunctional(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	// Migrate tracking models
	err = db.AutoMigrate(&tracking.TrackingEventModel{}, &tracking.ShipmentModel{})
	require.NoError(t, err)

	// Seed Dummy Shipment to update its status later
	resiID := "FUNC-RESI-001"
	db.Create(&tracking.ShipmentModel{
		ResiID:    resiID,
		Status:    domain.StatusCreated,
		UpdatedAt: time.Now(),
	})

	repo := tracking.NewTrackingRepository(db)
	service := tracking.NewTrackingService(repo)

	t.Run("Add Tracking Event Functional", func(t *testing.T) {
		err := service.AddTrackingEvent(resiID, domain.StatusInHub, "Hub JKT", "Arrived")
		assert.NoError(t, err)

		// Verify event was saved directly in DB
		var event tracking.TrackingEventModel
		err = db.Where("resi_id = ?", resiID).First(&event).Error
		assert.NoError(t, err)
		assert.Equal(t, domain.StatusInHub, event.Status)

		// Verify shipment was updated
		var shipment tracking.ShipmentModel
		err = db.Where("resi_id = ?", resiID).First(&shipment).Error
		assert.NoError(t, err)
		assert.Equal(t, domain.StatusInHub, shipment.Status)
	})

	t.Run("Get Tracking History Functional", func(t *testing.T) {
		err := service.AddTrackingEvent(resiID, domain.StatusInTransit, "Hub BDG", "Departed")
		assert.NoError(t, err)

		events, err := service.GetHistory(resiID)
		assert.NoError(t, err)
		assert.Len(t, events, 2)
		
		// Order must be chronological
		assert.Equal(t, domain.StatusInHub, events[0].Status)
		assert.Equal(t, domain.StatusInTransit, events[1].Status)
	})
}
