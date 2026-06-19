// Package postgres provides the concrete repository implementation for Hub and Courier services
// using PostgreSQL as the backing store via GORM.
//
// This package mirrors the interface of infrastructure/sqlite but uses PostgreSQL,
// making it suitable for production deployments. The infrastructure/sqlite package
// is retained for functional tests that rely on in-memory SQLite.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tubes-cc/logistics/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ShipmentModel represents the shipments table in PostgreSQL.
type ShipmentModel struct {
	ResiID        string                `gorm:"primaryKey;column:resi_id"`
	Status        domain.TrackingStatus `gorm:"column:status;not null;default:'CREATED'"`
	HubID         string                `gorm:"column:hub_id;not null;default:''"`
	CourierID     string                `gorm:"column:courier_id;not null;default:''"`
	CourierUserID string                `gorm:"column:courier_user_id;not null;default:''"`
	ProofURL      string                `gorm:"column:proof_url;not null;default:''"`
	UpdatedAt     time.Time             `gorm:"column:updated_at;not null"`
}

// TableName returns the explicit table name.
func (ShipmentModel) TableName() string {
	return "shipments"
}

// TrackingEventModel represents the tracking_events table in PostgreSQL.
type TrackingEventModel struct {
	ID        string                `gorm:"primaryKey;column:id"`
	ResiID    string                `gorm:"column:resi_id;not null"`
	Status    domain.TrackingStatus `gorm:"column:status;not null"`
	Location  string                `gorm:"column:location;not null;default:''"`
	Note      string                `gorm:"column:note;not null;default:''"`
	CreatedAt time.Time             `gorm:"column:created_at;not null"`
}

// TableName returns the explicit table name.
func (TrackingEventModel) TableName() string {
	return "tracking_events"
}

// ShipmentRepository implements hub.Repository and courier.Repository
// using PostgreSQL via GORM.
type ShipmentRepository struct {
	db *gorm.DB
}

// NewShipmentRepository creates a new repository and runs AutoMigrate.
func NewShipmentRepository(db *gorm.DB) (*ShipmentRepository, error) {
	repo := &ShipmentRepository{db: db}
	// if err := repo.migrate(); err != nil {
	// 	return nil, fmt.Errorf("failed to migrate PostgreSQL schema: %w", err)
	// }
	return repo, nil
}

// migrate creates all required tables using GORM AutoMigrate.
func (r *ShipmentRepository) migrate() error {
	if err := r.db.AutoMigrate(&ShipmentModel{}, &TrackingEventModel{}); err != nil {
		return err
	}
	return nil
}

func toDomainShipment(m *ShipmentModel) *domain.Shipment {
	if m == nil {
		return nil
	}
	return &domain.Shipment{
		ResiID:        m.ResiID,
		Status:        m.Status,
		HubID:         m.HubID,
		CourierID:     m.CourierID,
		CourierUserID: m.CourierUserID,
		ProofURL:      m.ProofURL,
		UpdatedAt:     m.UpdatedAt,
	}
}

func toDomainTrackingEvent(m *TrackingEventModel) *domain.TrackingEvent {
	if m == nil {
		return nil
	}
	return &domain.TrackingEvent{
		ID:        m.ID,
		ResiID:    m.ResiID,
		Status:    m.Status,
		Location:  m.Location,
		Note:      m.Note,
		CreatedAt: m.CreatedAt,
	}
}

// CreateShipmentIfNotExists creates a new shipment record idempotently.
func (r *ShipmentRepository) CreateShipmentIfNotExists(ctx context.Context, resiID string) error {
	m := &ShipmentModel{
		ResiID:    resiID,
		Status:    domain.StatusCreated,
		HubID:     "",
		CourierID: "",
		ProofURL:  "",
		UpdatedAt: time.Now(),
	}

	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(m).Error
	if err != nil {
		return fmt.Errorf("failed to create shipment: %w", err)
	}
	return nil
}

// GetShipment retrieves a shipment by resiID.
func (r *ShipmentRepository) GetShipment(ctx context.Context, resiID string) (*domain.Shipment, error) {
	var m ShipmentModel
	err := r.db.WithContext(ctx).Where("resi_id = ?", resiID).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrShipmentNotFound
		}
		return nil, fmt.Errorf("failed to query shipment: %w", err)
	}
	return toDomainShipment(&m), nil
}

// UpdateShipment persists changes to an existing shipment.
func (r *ShipmentRepository) UpdateShipment(ctx context.Context, shipment *domain.Shipment) error {
	updates := map[string]interface{}{
		"status":          string(shipment.Status),
		"hub_id":          shipment.HubID,
		"courier_id":      shipment.CourierID,
		"courier_user_id": shipment.CourierUserID,
		"proof_url":       shipment.ProofURL,
		"updated_at":      shipment.UpdatedAt,
	}

	result := r.db.WithContext(ctx).Model(&ShipmentModel{}).Where("resi_id = ?", shipment.ResiID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update shipment: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrShipmentNotFound
	}
	return nil
}

// AddTrackingEvent saves a tracking event.
func (r *ShipmentRepository) AddTrackingEvent(ctx context.Context, event *domain.TrackingEvent) error {
	m := &TrackingEventModel{
		ID:        event.ID,
		ResiID:    event.ResiID,
		Status:    event.Status,
		Location:  event.Location,
		Note:      event.Note,
		CreatedAt: event.CreatedAt,
	}

	err := r.db.WithContext(ctx).Create(m).Error
	if err != nil {
		return fmt.Errorf("failed to insert tracking event: %w", err)
	}
	return nil
}

// GetTrackingEvents returns all tracking events for a resi, ordered chronologically.
func (r *ShipmentRepository) GetTrackingEvents(ctx context.Context, resiID string) ([]*domain.TrackingEvent, error) {
	var models []TrackingEventModel
	err := r.db.WithContext(ctx).
		Where("resi_id = ?", resiID).
		Order("created_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query tracking events: %w", err)
	}

	events := make([]*domain.TrackingEvent, len(models))
	for i, m := range models {
		events[i] = toDomainTrackingEvent(&m)
	}
	return events, nil
}

// ListShipmentsByCourier returns all shipments assigned to a specific courier.
func (r *ShipmentRepository) ListShipmentsByCourier(ctx context.Context, courierUserID string) ([]*domain.Shipment, error) {
	var models []ShipmentModel
	err := r.db.WithContext(ctx).
		Where("courier_user_id = ? AND courier_user_id != ''", courierUserID).
		Order("updated_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query shipments by courier: %w", err)
	}

	shipments := make([]*domain.Shipment, len(models))
	for i, m := range models {
		shipments[i] = toDomainShipment(&m)
	}
	return shipments, nil
}