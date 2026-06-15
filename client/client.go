// Package client mendefinisikan interface untuk komunikasi antar-microservice.
//
// Penggunaan interface (bukan implementasi konkret) memungkinkan:
//   - Dependency injection
//   - Mock mudah di unit test
//   - Swap implementasi (HTTP, gRPC, dll) tanpa ubah service logic
package client

import (
	"context"

	"github.com/tubes-cc/logistics/domain"
)

// TrackingClient mendefinisikan kontrak untuk mengirim tracking event
// ke Tracking Service.
//
// Dalam production, implementasinya akan membuat HTTP request ke
// Tracking Service API. Dalam test, di-mock agar tidak butuh network.
type TrackingClient interface {
	// AddTrackingEvent mengirim event tracking baru ke Tracking Service.
	// Dipanggil setiap kali status paket berubah.
	AddTrackingEvent(ctx context.Context, event *domain.TrackingEvent) error

	// GetCurrentStatus retrieves the current shipment status from Tracking Service.
	// Returns domain.StatusCreated if no snapshot exists (fresh resi).
	GetCurrentStatus(ctx context.Context, resiID string) (domain.TrackingStatus, error)

	// GetTrackingHistory retrieves all tracking events for a resi from Tracking Service.
	// Returns empty slice (not nil) if no events exist (404 from Tracking).
	GetTrackingHistory(ctx context.Context, resiID string) ([]domain.TrackingEvent, error)
}

// PricingClient mendefinisikan kontrak untuk meminta perhitungan harga
// ke Pricing Service.
type PricingClient interface {
	// CalculatePrice meminta Pricing Service menghitung ongkos kirim.
	CalculatePrice(ctx context.Context, req domain.PricingRequest) (*domain.PricingResult, error)
}

// OrderClient mendefinisikan kontrak untuk validasi resi
// ke Order Service.
//
// Memastikan bahwa resi yang diproses benar-benar ada dan valid
// sebelum melakukan operasi apa pun.
type OrderClient interface {
	// ValidateResi memvalidasi apakah resiID ada dan aktif di Order Service.
	// Mengembalikan domain.ErrResiInvalid jika resi tidak valid.
	ValidateResi(ctx context.Context, resiID string) error
}

// AuthClient mendefinisikan kontrak untuk validasi token JWT
// ke Auth Service.
//
// Digunakan oleh middleware untuk mengautentikasi setiap request.
type AuthClient interface {
	// ValidateToken memvalidasi token JWT dan mengembalikan klaim-nya.
	// Mengembalikan domain.ErrUnauthorized jika token tidak valid.
	ValidateToken(ctx context.Context, token string) (*domain.AuthClaims, error)
}
