package tracking

import (
	"errors"
	"time"

	"github.com/tubes-cc/logistics/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TrackingEventModel is the DB schema for tracking events.
type TrackingEventModel struct {
	ID        string `gorm:"primaryKey"`
	ResiID    string `gorm:"index"`
	Status    domain.TrackingStatus
	Location  string
	Note      string
	CreatedAt time.Time
}

// ShipmentModel is the DB schema for shipments.
type ShipmentModel struct {
	ResiID    string `gorm:"primaryKey"`
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
	GetCurrentStatus(resiID string) (domain.TrackingStatus, error)
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

// UpdateShipmentStatus updates the latest status of a shipment using upsert to guarantee existence.
func (r *trackingRepository) UpdateShipmentStatus(resiID string, status domain.TrackingStatus) error {
	shipment := ShipmentModel{
		ResiID:    resiID,
		Status:    status,
		UpdatedAt: time.Now(),
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "resi_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "updated_at"}),
	}).Create(&shipment).Error
}

// GetCurrentStatus returns the current shipment status from the snapshot table.
// Returns domain.ErrShipmentNotFound if no row exists for the given resiID.
func (r *trackingRepository) GetCurrentStatus(resiID string) (domain.TrackingStatus, error) {
	var shipment ShipmentModel
	err := r.db.Where("resi_id = ?", resiID).First(&shipment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", domain.ErrShipmentNotFound
		}
		return "", err
	}
	return shipment.Status, nil
}
