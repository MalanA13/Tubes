// Package sqlite menyediakan implementasi konkret dari repository
// menggunakan SQLite. Package ini berada di outermost layer (infrastructure)
// dan mengimplementasikan interface dari hub.Repository dan courier.Repository.
package sqlite

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tubes-cc/logistics/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ShipmentModel merepresentasikan tabel shipments di database.
type ShipmentModel struct {
	ResiID    string                `gorm:"primaryKey;column:resi_id"`
	Status    domain.TrackingStatus `gorm:"column:status;not null;default:'CREATED'"`
	HubID     string                `gorm:"column:hub_id;not null;default:''"`
	CourierID string                `gorm:"column:courier_id;not null;default:''"`
	ProofURL  string                `gorm:"column:proof_url;not null;default:''"`
	UpdatedAt time.Time             `gorm:"column:updated_at;not null"`
}

// TableName menentukan nama tabel shipments secara eksplisit.
func (ShipmentModel) TableName() string {
	return "shipments"
}

// TrackingEventModel merepresentasikan tabel tracking_events di database.
type TrackingEventModel struct {
	ID        string                `gorm:"primaryKey;column:id"`
	ResiID    string                `gorm:"column:resi_id;not null"`
	Status    domain.TrackingStatus `gorm:"column:status;not null"`
	Location  string                `gorm:"column:location;not null;default:''"`
	Note      string                `gorm:"column:note;not null;default:''"`
	CreatedAt time.Time             `gorm:"column:created_at;not null"`
}

// TableName menentukan nama tabel tracking_events secara eksplisit.
func (TrackingEventModel) TableName() string {
	return "tracking_events"
}

// ShipmentRepository mengimplementasikan hub.Repository dan courier.Repository
// menggunakan SQLite sebagai penyimpanan data dengan GORM.
type ShipmentRepository struct {
	db *gorm.DB
}

// NewShipmentRepository membuat instance repository baru dan menjalankan migrasi.
// Parameter db adalah koneksi GORM SQLite yang sudah terbuka.
func NewShipmentRepository(db *gorm.DB) (*ShipmentRepository, error) {
	repo := &ShipmentRepository{db: db}
	if err := repo.migrate(); err != nil {
		return nil, fmt.Errorf("gagal migrasi database: %w", err)
	}
	return repo, nil
}

// migrate membuat semua tabel yang diperlukan jika belum ada menggunakan GORM AutoMigrate.
func (r *ShipmentRepository) migrate() error {
	if err := r.db.AutoMigrate(&ShipmentModel{}, &TrackingEventModel{}); err != nil {
		return err
	}
	return nil
}

// toDomainShipment mengubah ShipmentModel menjadi domain.Shipment.
func toDomainShipment(m *ShipmentModel) *domain.Shipment {
	if m == nil {
		return nil
	}
	return &domain.Shipment{
		ResiID:    m.ResiID,
		Status:    m.Status,
		HubID:     m.HubID,
		CourierID: m.CourierID,
		ProofURL:  m.ProofURL,
		UpdatedAt: m.UpdatedAt,
	}
}

// toDomainTrackingEvent mengubah TrackingEventModel menjadi domain.TrackingEvent.
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

// CreateShipmentIfNotExists membuat record shipment baru jika resiID belum ada.
// Jika sudah ada, operasi ini diabaikan (idempotent) menggunakan clauses on conflict do nothing.
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
		return fmt.Errorf("gagal create shipment: %w", err)
	}
	return nil
}

// GetShipment mengambil data shipment berdasarkan resiID.
// Mengembalikan domain.ErrShipmentNotFound jika tidak ada.
func (r *ShipmentRepository) GetShipment(ctx context.Context, resiID string) (*domain.Shipment, error) {
	var m ShipmentModel
	err := r.db.WithContext(ctx).Where("resi_id = ?", resiID).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrShipmentNotFound
		}
		return nil, fmt.Errorf("gagal query shipment: %w", err)
	}
	return toDomainShipment(&m), nil
}

// UpdateShipment menyimpan perubahan data shipment ke database.
// Mengembalikan domain.ErrShipmentNotFound jika resiID tidak ada.
func (r *ShipmentRepository) UpdateShipment(ctx context.Context, shipment *domain.Shipment) error {
	updates := map[string]interface{}{
		"status":     string(shipment.Status),
		"hub_id":     shipment.HubID,
		"courier_id": shipment.CourierID,
		"proof_url":  shipment.ProofURL,
		"updated_at": shipment.UpdatedAt,
	}

	result := r.db.WithContext(ctx).Model(&ShipmentModel{}).Where("resi_id = ?", shipment.ResiID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("gagal update shipment: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrShipmentNotFound
	}
	return nil
}

// AddTrackingEvent menyimpan tracking event ke database.
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
		return fmt.Errorf("gagal insert tracking event: %w", err)
	}
	return nil
}

// GetTrackingEvents mengambil semua tracking event untuk sebuah resi,
// diurutkan berdasarkan waktu (ascending = kronologis).
func (r *ShipmentRepository) GetTrackingEvents(ctx context.Context, resiID string) ([]*domain.TrackingEvent, error) {
	var models []TrackingEventModel
	err := r.db.WithContext(ctx).
		Where("resi_id = ?", resiID).
		Order("created_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("gagal query tracking events: %w", err)
	}

	events := make([]*domain.TrackingEvent, len(models))
	for i, m := range models {
		events[i] = toDomainTrackingEvent(&m)
	}
	return events, nil
}

// Open membuka koneksi SQLite menggunakan GORM. Gunakan ":memory:" untuk in-memory database.
func Open(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("gagal membuka SQLite: %w", err)
	}
	
	// Verifikasi koneksi dengan ping underlying DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan sql.DB: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("gagal ping SQLite: %w", err)
	}
	
	return db, nil
}
