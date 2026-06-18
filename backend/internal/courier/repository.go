// Package courier — repository interface untuk Courier Service.
package courier

import (
	"context"

	"github.com/tubes-cc/logistics/domain"
)

// Repository mendefinisikan kontrak akses data untuk Courier Service.
type Repository interface {
	// GetShipment mengambil data shipment berdasarkan resi ID.
	GetShipment(ctx context.Context, resiID string) (*domain.Shipment, error)

	// UpdateShipment menyimpan perubahan data shipment.
	UpdateShipment(ctx context.Context, shipment *domain.Shipment) error

	// CreateShipmentIfNotExists membuat shipment baru jika belum ada.
	CreateShipmentIfNotExists(ctx context.Context, resiID string) error

	// ListShipmentsByCourier mengambil semua shipment yang ditugaskan ke courier tertentu.
	ListShipmentsByCourier(ctx context.Context, courierUserID string) ([]*domain.Shipment, error)
}
