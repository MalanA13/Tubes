// Package courier — business logic untuk Courier Service.
// Bergantung pada: Repository, client.TrackingClient, client.OrderClient.
package courier

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tubes-cc/logistics/client"
	"github.com/tubes-cc/logistics/domain"
)

// Service menyediakan operasi bisnis Courier Service.
type Service struct {
	repo           Repository
	trackingClient client.TrackingClient
	orderClient    client.OrderClient
}

// NewService membuat instance Courier Service baru.
//
// Contoh:
//   svc := courier.NewService(sqliteRepo, httpTrackingClient, httpOrderClient)
func NewService(repo Repository, tc client.TrackingClient, oc client.OrderClient) *Service {
	return &Service{
		repo:           repo,
		trackingClient: tc,
		orderClient:    oc,
	}
}

// AssignCourier menugaskan kurir ke sebuah pengiriman.
//
// Flow:
//  1. Validasi resiID dan courierID
//  2. Validasi resi ke Order Service
//  3. Inisialisasi shipment jika belum ada
//  4. Status harus IN_TRANSIT; kurir belum boleh ditugaskan
//  5. Update status → OUT_DELIVERY
//  6. Kirim tracking event
func (s *Service) AssignCourier(ctx context.Context, resiID, courierID string) error {
	// 1. Validasi input
	if resiID == "" {
		return domain.ErrInvalidResiID
	}
	if courierID == "" {
		return domain.ErrInvalidCourierID
	}

	// 2. Validasi resi
	if err := s.orderClient.ValidateResi(ctx, resiID); err != nil {
		return fmt.Errorf("assign-courier: validasi resi gagal: %w", err)
	}

	// 3. Inisialisasi jika belum ada
	if err := s.repo.CreateShipmentIfNotExists(ctx, resiID); err != nil {
		return fmt.Errorf("assign-courier: inisialisasi shipment gagal: %w", err)
	}

	// 4. Ambil & validasi
	shipment, err := s.repo.GetShipment(ctx, resiID)
	if err != nil {
		return fmt.Errorf("assign-courier: gagal ambil shipment: %w", err)
	}

	if shipment.Status != domain.StatusInTransit {
		return fmt.Errorf("assign-courier: status saat ini '%s', harus 'IN_TRANSIT': %w",
			shipment.Status, domain.ErrInvalidStatus)
	}

	if shipment.CourierID != "" {
		return domain.ErrCourierAlreadyAssigned
	}

	// 5. Update
	now := time.Now()
	shipment.CourierID = courierID
	shipment.Status = domain.StatusOutDelivery
	shipment.UpdatedAt = now

	if err := s.repo.UpdateShipment(ctx, shipment); err != nil {
		return fmt.Errorf("assign-courier: gagal update shipment: %w", err)
	}

	// 6. Kirim tracking event dengan UUID
	event := &domain.TrackingEvent{
		ID:        uuid.New().String(),
		ResiID:    resiID,
		Status:    domain.StatusOutDelivery,
		Location:  fmt.Sprintf("courier-%s", courierID),
		Note:      fmt.Sprintf("Paket ditugaskan ke kurir %s", courierID),
		CreatedAt: now,
	}

	if err := s.trackingClient.AddTrackingEvent(ctx, event); err != nil {
		return fmt.Errorf("assign-courier: gagal kirim tracking event: %w", err)
	}

	return nil
}

// validDeliveryStatuses adalah set status yang diizinkan untuk update delivery.
var validDeliveryStatuses = map[domain.TrackingStatus]bool{
	domain.StatusDelivered: true,
	domain.StatusFailed:    true,
	domain.StatusReturned:  true,
}

// UpdateDeliveryStatus memperbarui status pengiriman akhir oleh kurir.
//
// Flow:
//  1. Validasi resiID
//  2. Validasi status (DELIVERED/FAILED/RETURNED) dan proofURL
//  3. Validasi resi ke Order Service
//  4. Status harus OUT_DELIVERY
//  5. Update status & proof
//  6. Kirim tracking event
func (s *Service) UpdateDeliveryStatus(ctx context.Context, resiID string, status domain.TrackingStatus, proofURL string) error {
	// 1. Validasi input dasar
	if resiID == "" {
		return domain.ErrInvalidResiID
	}

	// 2. Validasi status dan proof URL
	if !validDeliveryStatuses[status] {
		return fmt.Errorf("update-delivery: status '%s' tidak valid (gunakan DELIVERED/FAILED/RETURNED): %w",
			status, domain.ErrInvalidStatus)
	}
	if status == domain.StatusDelivered && proofURL == "" {
		return domain.ErrInvalidProofURL
	}

	// 3. Validasi resi
	if err := s.orderClient.ValidateResi(ctx, resiID); err != nil {
		return fmt.Errorf("update-delivery: validasi resi gagal: %w", err)
	}

	// 4. Ambil & validasi
	shipment, err := s.repo.GetShipment(ctx, resiID)
	if err != nil {
		return fmt.Errorf("update-delivery: gagal ambil shipment: %w", err)
	}

	if shipment.Status != domain.StatusOutDelivery {
		return fmt.Errorf("update-delivery: status saat ini '%s', harus 'OUT_DELIVERY': %w",
			shipment.Status, domain.ErrInvalidStatus)
	}

	// 5. Update
	now := time.Now()
	shipment.Status = status
	shipment.ProofURL = proofURL
	shipment.UpdatedAt = now

	if err := s.repo.UpdateShipment(ctx, shipment); err != nil {
		return fmt.Errorf("update-delivery: gagal update shipment: %w", err)
	}

	// 6. Kirim tracking event dengan UUID
	note := fmt.Sprintf("Status pengiriman diperbarui: %s", status)
	if proofURL != "" {
		note += fmt.Sprintf(" | bukti: %s", proofURL)
	}

	event := &domain.TrackingEvent{
		ID:        uuid.New().String(),
		ResiID:    resiID,
		Status:    status,
		Location:  fmt.Sprintf("courier-%s", shipment.CourierID),
		Note:      note,
		CreatedAt: now,
	}

	if err := s.trackingClient.AddTrackingEvent(ctx, event); err != nil {
		return fmt.Errorf("update-delivery: gagal kirim tracking event: %w", err)
	}

	return nil
}
