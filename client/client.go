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
