package functional

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tubes-cc/logistics/domain"
	"github.com/tubes-cc/logistics/internal/handler"
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

	repo := tracking.NewTrackingRepository(db)
	service := tracking.NewTrackingService(repo)

	// Real production flow verification: NO manual seeding of ShipmentModel.
	resiID := "FUNC-RESI-AUTO-001"

	t.Run("First Event Creates Shipment Functional", func(t *testing.T) {
		err := service.AddTrackingEvent(resiID, domain.StatusInHub, "Hub JKT", "Arrived")
		assert.NoError(t, err)

		// Verify event was saved directly in DB
		var event tracking.TrackingEventModel
		err = db.Where("resi_id = ?", resiID).First(&event).Error
		assert.NoError(t, err)
		assert.Equal(t, domain.StatusInHub, event.Status)

		// Verify shipment was automatically created via upsert
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

		// Verify shipment status was updated to latest event
		var shipment tracking.ShipmentModel
		err = db.Where("resi_id = ?", resiID).First(&shipment).Error
		assert.NoError(t, err)
		assert.Equal(t, domain.StatusInTransit, shipment.Status)
	})
}

func TestTrackingConsistency(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&tracking.TrackingEventModel{}, &tracking.ShipmentModel{})
	require.NoError(t, err)

	repo := tracking.NewTrackingRepository(db)
	service := tracking.NewTrackingService(repo)

	resiID := "CONSISTENCY-RESI-123"

	// 1. Verify first event creates shipment
	err = service.AddTrackingEvent(resiID, domain.StatusInHub, "Hub A", "Scanned In")
	require.NoError(t, err)

	var shipment tracking.ShipmentModel
	err = db.Where("resi_id = ?", resiID).First(&shipment).Error
	require.NoError(t, err)
	assert.Equal(t, domain.StatusInHub, shipment.Status)
	// Other fields are zero values
	assert.Empty(t, shipment.HubID)
	assert.Empty(t, shipment.CourierID)
	assert.Empty(t, shipment.ProofURL)

	// Simulate external service manually writing Courier ID (e.g. Courier Svc assign)
	db.Model(&tracking.ShipmentModel{}).Where("resi_id = ?", resiID).Update("courier_id", "COURIER-456")

	// 2. Verify second event updates shipment and preserves CourierID
	err = service.AddTrackingEvent(resiID, domain.StatusInTransit, "Hub B", "Scanned Out")
	require.NoError(t, err)

	err = db.Where("resi_id = ?", resiID).First(&shipment).Error
	require.NoError(t, err)
	assert.Equal(t, domain.StatusInTransit, shipment.Status)
	assert.Equal(t, "COURIER-456", shipment.CourierID) // Preserved!

	// 3. Verify shipment status matches latest tracking event
	events, err := service.GetHistory(resiID)
	require.NoError(t, err)
	require.Len(t, events, 2)
	assert.Equal(t, events[len(events)-1].Status, shipment.Status)
}

func TestTrackingHTTPHandler(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&tracking.TrackingEventModel{}, &tracking.ShipmentModel{})
	require.NoError(t, err)

	repo := tracking.NewTrackingRepository(db)
	service := tracking.NewTrackingService(repo)

	// Register Gorilla Mux Router
	r := mux.NewRouter()
	r.HandleFunc("/tracking/{resiID}", handler.HandleGetTrackingHTTP(*service)).Methods("GET")

	// 1. Test 404 for non-existent resi
	req := httptest.NewRequest(http.MethodGet, "/tracking/NON-EXISTENT", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	// 2. Test 200 OK after adding event
	resiID := "HTTP-RESI-789"
	err = service.AddTrackingEvent(resiID, domain.StatusCreated, "Warehouse", "Order Created")
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodGet, "/tracking/"+resiID, nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var respBody struct {
		Success bool                    `json:"success"`
		Data    []domain.TrackingEvent `json:"data"`
	}
	err = json.NewDecoder(rec.Body).Decode(&respBody)
	require.NoError(t, err)
	assert.True(t, respBody.Success)
	require.Len(t, respBody.Data, 1)
	assert.Equal(t, domain.StatusCreated, respBody.Data[0].Status)
	assert.Equal(t, "Warehouse", respBody.Data[0].Location)
}
