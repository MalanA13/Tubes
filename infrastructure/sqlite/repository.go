// Package sqlite menyediakan implementasi konkret dari repository
// menggunakan SQLite. Package ini berada di outermost layer (infrastructure)
// dan mengimplementasikan interface dari hub.Repository dan courier.Repository.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/tubes-cc/logistics/domain"

	// Driver SQLite — hanya diperlukan side-effect (registrasi driver).
	_ "github.com/mattn/go-sqlite3"
)

// ShipmentRepository mengimplementasikan hub.Repository dan courier.Repository
// menggunakan SQLite sebagai penyimpanan data.
type ShipmentRepository struct {
	db *sql.DB
}

// NewShipmentRepository membuat instance repository baru dan menjalankan migrasi.
// Parameter db adalah koneksi SQLite yang sudah terbuka.
//
// Contoh untuk in-memory (test):
//   db, _ := sql.Open("sqlite3", ":memory:")
//   repo, _ := sqlite.NewShipmentRepository(db)
//
// Contoh untuk file (production):
//   db, _ := sql.Open("sqlite3", "./logistics.db")
//   repo, _ := sqlite.NewShipmentRepository(db)
func NewShipmentRepository(db *sql.DB) (*ShipmentRepository, error) {
	repo := &ShipmentRepository{db: db}
	if err := repo.migrate(); err != nil {
		return nil, fmt.Errorf("gagal migrasi database: %w", err)
	}
	return repo, nil
}

// migrate membuat semua tabel yang diperlukan jika belum ada.
func (r *ShipmentRepository) migrate() error {
	queries := []string{
		// Tabel shipments: data utama pengiriman
		`CREATE TABLE IF NOT EXISTS shipments (
			resi_id    TEXT PRIMARY KEY,
			status     TEXT NOT NULL DEFAULT 'CREATED',
			hub_id     TEXT NOT NULL DEFAULT '',
			courier_id TEXT NOT NULL DEFAULT '',
			proof_url  TEXT NOT NULL DEFAULT '',
			updated_at TEXT NOT NULL DEFAULT ''
		)`,
		// Tabel tracking_events: history perubahan status
		// ID berupa UUID sehingga PRIMARY KEY tidak akan collision
		`CREATE TABLE IF NOT EXISTS tracking_events (
			id         TEXT PRIMARY KEY,
			resi_id    TEXT NOT NULL,
			status     TEXT NOT NULL,
			location   TEXT NOT NULL DEFAULT '',
			note       TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL DEFAULT ''
		)`,
	}

	for _, q := range queries {
		if _, err := r.db.Exec(q); err != nil {
			return fmt.Errorf("gagal eksekusi migrasi: %w", err)
		}
	}
	return nil
}

// CreateShipmentIfNotExists membuat record shipment baru jika resiID belum ada.
// Jika sudah ada, operasi ini diabaikan (idempotent).
// Dipanggil pada awal ScanIn dan AssignCourier untuk memastikan record tersedia.
func (r *ShipmentRepository) CreateShipmentIfNotExists(ctx context.Context, resiID string) error {
	query := `INSERT OR IGNORE INTO shipments (resi_id, status, hub_id, courier_id, proof_url, updated_at)
	          VALUES (?, 'CREATED', '', '', '', ?)`

	_, err := r.db.ExecContext(ctx, query, resiID, time.Now().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("gagal create shipment: %w", err)
	}
	return nil
}

// GetShipment mengambil data shipment berdasarkan resiID.
// Mengembalikan domain.ErrShipmentNotFound jika tidak ada.
func (r *ShipmentRepository) GetShipment(ctx context.Context, resiID string) (*domain.Shipment, error) {
	query := `SELECT resi_id, status, hub_id, courier_id, proof_url, updated_at
	          FROM shipments WHERE resi_id = ?`

	var s domain.Shipment
	var updatedAtStr string

	err := r.db.QueryRowContext(ctx, query, resiID).Scan(
		&s.ResiID, &s.Status, &s.HubID, &s.CourierID, &s.ProofURL, &updatedAtStr,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrShipmentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("gagal query shipment: %w", err)
	}

	// Parse waktu — toleran terhadap berbagai format SQLite
	s.UpdatedAt = parseTime(updatedAtStr)
	return &s, nil
}

// UpdateShipment menyimpan perubahan data shipment ke database.
// Mengembalikan domain.ErrShipmentNotFound jika resiID tidak ada.
func (r *ShipmentRepository) UpdateShipment(ctx context.Context, shipment *domain.Shipment) error {
	query := `UPDATE shipments
	          SET status = ?, hub_id = ?, courier_id = ?, proof_url = ?, updated_at = ?
	          WHERE resi_id = ?`

	result, err := r.db.ExecContext(ctx, query,
		string(shipment.Status),
		shipment.HubID,
		shipment.CourierID,
		shipment.ProofURL,
		shipment.UpdatedAt.Format(time.RFC3339Nano),
		shipment.ResiID,
	)
	if err != nil {
		return fmt.Errorf("gagal update shipment: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrShipmentNotFound
	}
	return nil
}

// AddTrackingEvent menyimpan tracking event ke database.
// Event ID harus berupa UUID agar tidak terjadi UNIQUE constraint violation.
func (r *ShipmentRepository) AddTrackingEvent(ctx context.Context, event *domain.TrackingEvent) error {
	query := `INSERT INTO tracking_events (id, resi_id, status, location, note, created_at)
	          VALUES (?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		event.ID,
		event.ResiID,
		string(event.Status),
		event.Location,
		event.Note,
		event.CreatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("gagal insert tracking event: %w", err)
	}
	return nil
}

// GetTrackingEvents mengambil semua tracking event untuk sebuah resi,
// diurutkan berdasarkan waktu (ascending = kronologis).
// Digunakan oleh functional test untuk verifikasi history.
func (r *ShipmentRepository) GetTrackingEvents(ctx context.Context, resiID string) ([]*domain.TrackingEvent, error) {
	query := `SELECT id, resi_id, status, location, note, created_at
	          FROM tracking_events
	          WHERE resi_id = ?
	          ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, resiID)
	if err != nil {
		return nil, fmt.Errorf("gagal query tracking events: %w", err)
	}
	defer rows.Close()

	var events []*domain.TrackingEvent
	for rows.Next() {
		var e domain.TrackingEvent
		var createdAtStr string

		if err := rows.Scan(&e.ID, &e.ResiID, &e.Status, &e.Location, &e.Note, &createdAtStr); err != nil {
			return nil, fmt.Errorf("gagal scan event: %w", err)
		}
		e.CreatedAt = parseTime(createdAtStr)
		events = append(events, &e)
	}
	return events, rows.Err()
}

// Open membuka koneksi SQLite. Gunakan ":memory:" untuk in-memory database.
func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka SQLite: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("gagal ping SQLite: %w", err)
	}
	return db, nil
}

// parseTime memparse string waktu dari SQLite ke time.Time.
// Toleran terhadap beberapa format yang mungkin dihasilkan SQLite.
func parseTime(s string) time.Time {
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Time{}
}
