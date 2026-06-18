package functional_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tubes-cc/logistics/domain"
	"github.com/tubes-cc/logistics/internal/handler"
	"github.com/tubes-cc/logistics/internal/notification"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupNotificationDB creates a fresh in-memory database with the notifications table.
func setupNotificationDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create in-memory SQLite database: %v", err)
	}
	if err := db.AutoMigrate(&domain.Notification{}); err != nil {
		t.Fatalf("Failed to auto-migrate: %v", err)
	}
	return db
}

// ---------- Service-level functional tests ----------

func TestNotification_Functional(t *testing.T) {
	db := setupNotificationDB(t)

	notifRepo := notification.NewNotificationRepository(db)
	notifService := notification.NewNotificationService(notifRepo)

	// Send a notification via the service
	err := notifService.CreateNotification("user-007", "Paketmu sedang dikirim kurir!", "SMS")
	assert.NoError(t, err)

	// Verify it was persisted in the database
	var savedNotif domain.Notification
	result := db.First(&savedNotif, "user_id = ?", "user-007")
	assert.NoError(t, result.Error, "Notification should be saved in DB")
	assert.Equal(t, "SMS", savedNotif.Type)
	assert.Equal(t, "Paketmu sedang dikirim kurir!", savedNotif.Message)
	assert.Equal(t, "SENT", savedNotif.Status)
}

func TestNotification_EmptyMessage_Functional(t *testing.T) {
	db := setupNotificationDB(t)

	notifRepo := notification.NewNotificationRepository(db)
	notifService := notification.NewNotificationService(notifRepo)

	err := notifService.CreateNotification("user-007", "", "SMS")
	assert.Error(t, err, "Empty message should be rejected")
	assert.Equal(t, "pesan tidak boleh kosong", err.Error())
}

func TestNotification_EmailType_Functional(t *testing.T) {
	db := setupNotificationDB(t)

	notifRepo := notification.NewNotificationRepository(db)
	notifService := notification.NewNotificationService(notifRepo)

	err := notifService.CreateNotification("user-123", "Order confirmed!", "EMAIL")
	assert.NoError(t, err)

	var savedNotif domain.Notification
	result := db.First(&savedNotif, "user_id = ?", "user-123")
	assert.NoError(t, result.Error)
	assert.Equal(t, "EMAIL", savedNotif.Type)
	assert.Equal(t, "Order confirmed!", savedNotif.Message)
}

// ---------- HTTP handler functional tests ----------

func TestNotifyHTTP_Success_Functional(t *testing.T) {
	db := setupNotificationDB(t)

	notifRepo := notification.NewNotificationRepository(db)
	notifService := notification.NewNotificationService(notifRepo)

	h := handler.HandleSendNotificationHTTP(*notifService)

	body, _ := json.Marshal(map[string]string{
		"user_id": "user-http-1",
		"message": "Paket dikirim via API!",
		"type":    "SMS",
	})

	req := httptest.NewRequest(http.MethodPost, "/notify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code, "Should return 201 Created")

	var resp struct {
		Success bool              `json:"success"`
		Data    map[string]string `json:"data"`
	}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "sent", resp.Data["status"])

	// Verify persistence
	var savedNotif domain.Notification
	result := db.First(&savedNotif, "user_id = ?", "user-http-1")
	assert.NoError(t, result.Error)
	assert.Equal(t, "Paket dikirim via API!", savedNotif.Message)
}

func TestNotifyHTTP_MissingFields_Functional(t *testing.T) {
	db := setupNotificationDB(t)

	notifRepo := notification.NewNotificationRepository(db)
	notifService := notification.NewNotificationService(notifRepo)

	h := handler.HandleSendNotificationHTTP(*notifService)

	// Missing "type" field
	body, _ := json.Marshal(map[string]string{
		"user_id": "user-http-2",
		"message": "No type provided",
	})

	req := httptest.NewRequest(http.MethodPost, "/notify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code, "Should return 400 for missing fields")
}

func TestNotifyHTTP_InvalidBody_Functional(t *testing.T) {
	db := setupNotificationDB(t)

	notifRepo := notification.NewNotificationRepository(db)
	notifService := notification.NewNotificationService(notifRepo)

	h := handler.HandleSendNotificationHTTP(*notifService)

	req := httptest.NewRequest(http.MethodPost, "/notify", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code, "Should return 400 for invalid JSON")
}