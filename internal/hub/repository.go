// Package hub — repository interface untuk Hub Service.
// Interface didefinisikan di sini (layer service/use-case),
// bukan di layer infrastructure.
package hub

import (
	"context"

	"github.com/tubes-cc/logistics/domain"
)

// Repository mendefinisikan kontrak akses data untuk Hub Service.
// Implementasi konkret ada di infrastructure/sqlite/.
type Repository interface {
	// GetShipment mengambil data shipment berdasarkan resi ID.
	// Mengembalikan domain.ErrShipmentNotFound jika tidak ditemukan.
	GetShipment(ctx context.Context, resiID string) (*domain.Shipment, error)

	// UpdateShipment menyimpan perubahan data shipment.
	UpdateShipment(ctx context.Context, shipment *domain.Shipment) error

	// CreateShipmentIfNotExists membuat shipment baru jika belum ada.
	// Digunakan untuk inisialisasi pertama kali ketika resi valid tapi belum ada di DB hub.
	CreateShipmentIfNotExists(ctx context.Context, resiID string) error
}
