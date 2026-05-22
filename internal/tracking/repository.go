package tracking

import (
	"time"
	"github.com/tubes-cc/logistics/domain"
	"gorm.io/gorm"
)

// TrackingEventModel is the DB schema for tracking events.
type TrackingEventModel struct {
	ID        string                `gorm:"primaryKey"`
	ResiID    string                `gorm:"index"`
	Status    domain.TrackingStatus
	Location  string
	Note      string
	CreatedAt time.Time
}

// ShipmentModel is the DB schema for shipments.
type ShipmentModel struct {
	ResiID    string                `gorm:"primaryKey"`
	Status    domain.TrackingStatus
	HubID     string
	CourierID string
	ProofURL  string
	UpdatedAt time.Time
}

// TrackingRepository defines the interface for tracking-related DB operations
type TrackingRepository interface {
	SaveEvent(event domain.TrackingEvent) error
	GetTrackingHistory(resiID string) ([]domain.TrackingEvent, error)
	UpdateShipmentStatus(resiID string, status domain.TrackingStatus) error
}

// trackingRepository is the implementation of TrackingRepository using GORM
type trackingRepository struct {
	db *gorm.DB
}

// NewTrackingRepository creates a new instance of TrackingRepository
func NewTrackingRepository(db *gorm.DB) TrackingRepository {
	return &trackingRepository{db: db}
}

// SaveEvent saves a tracking event.
func (r *trackingRepository) SaveEvent(event domain.TrackingEvent) error {
	model := TrackingEventModel{
		ID:        event.ID,
		ResiID:    event.ResiID,
		Status:    event.Status,
		Location:  event.Location,
		Note:      event.Note,
		CreatedAt: event.CreatedAt,
	}
	return r.db.Create(&model).Error
}

// GetTrackingHistory fetches tracking history for a resi ID.
func (r *trackingRepository) GetTrackingHistory(resiID string) ([]domain.TrackingEvent, error) {
	var models []TrackingEventModel
	if err := r.db.Where("resi_id = ?", resiID).Order("created_at asc").Find(&models).Error; err != nil {
		return nil, err
	}
	var events []domain.TrackingEvent
	for _, m := range models {
		events = append(events, domain.TrackingEvent{
			ID:        m.ID,
			ResiID:    m.ResiID,
			Status:    m.Status,
			Location:  m.Location,
			Note:      m.Note,
			CreatedAt: m.CreatedAt,
		})
	}
	return events, nil
}

// UpdateShipmentStatus updates the latest status of a shipment.
func (r *trackingRepository) UpdateShipmentStatus(resiID string, status domain.TrackingStatus) error {
	return r.db.Model(&ShipmentModel{}).Where("resi_id = ?", resiID).Updates(map[string]interface{}{"status": status, "updated_at": time.Now()}).Error
}
