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
	authClient     client.AuthClient
}

// NewService membuat instance Courier Service baru.
//
// Contoh:
//   svc := courier.NewService(sqliteRepo, httpTrackingClient, httpOrderClient, httpAuthClient)
func NewService(repo Repository, tc client.TrackingClient, oc client.OrderClient, ac client.AuthClient) *Service {
	return &Service{
		repo:           repo,
		trackingClient: tc,
		orderClient:    oc,
		authClient:     ac,
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

	// 2. Validasi bahwa courierID merupakan User.ID yang valid dengan role='courier'
	// CourierID sekarang harus berupa User.ID (string dari uint, misal "123")
	if err := s.authClient.ValidateUserRole(ctx, courierID, domain.RoleCourier); err != nil {
		return fmt.Errorf("assign-courier: %w", err)
	}

	// 3. Validasi resi
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

	// Karena Hub dan Courier menggunakan DB terpisah, resi yang baru masuk ke DB Courier
	// akan memiliki status CREATED secara default dari CreateShipmentIfNotExists.
	// Oleh karena itu, kita izinkan status CREATED atau IN_TRANSIT.
	if shipment.Status != domain.StatusInTransit && shipment.Status != domain.StatusCreated {
		return fmt.Errorf("assign-courier: status saat ini '%s', harus 'IN_TRANSIT' atau 'CREATED': %w",
			shipment.Status, domain.ErrInvalidStatus)
	}

	if shipment.CourierID != "" {
		return domain.ErrCourierAlreadyAssigned
	}

	// 5. Update - store both CourierID and CourierUserID
	now := time.Now()
	shipment.CourierID = courierID           // Display identifier (sama dengan UserID untuk sistem baru)
	shipment.CourierUserID = courierID       // User.ID untuk ownership validation
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
//  1. Validasi resiID dan claims
//  2. Validasi status (DELIVERED/FAILED/RETURNED) dan proofURL
//  3. Validasi resi ke Order Service
//  4. Ambil shipment dan validasi ownership
//  5. Status harus OUT_DELIVERY
//  6. Update status & proof
//  7. Kirim tracking event
func (s *Service) UpdateDeliveryStatus(ctx context.Context, claims *domain.AuthClaims, resiID string, status domain.TrackingStatus, proofURL string) error {
	// 1. Validasi input dasar
	if claims == nil {
		return domain.ErrUnauthorized
	}
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

	// 4. Ambil shipment
	shipment, err := s.repo.GetShipment(ctx, resiID)
	if err != nil {
		return fmt.Errorf("update-delivery: gagal ambil shipment: %w", err)
	}

	// 5. OWNERSHIP VALIDATION - Critical security check
	// Shipment harus memiliki CourierUserID yang terisi (sistem baru)
	if shipment.CourierUserID == "" {
		return fmt.Errorf("update-delivery: shipment belum di-assign dengan sistem baru, tidak dapat diupdate: %w",
			domain.ErrUnauthorized)
	}
	
	// Validasi bahwa courier yang authenticated adalah pemilik shipment ini
	if claims.UserID != shipment.CourierUserID {
		return fmt.Errorf("update-delivery: %w", domain.ErrNotShipmentOwner)
	}

	// 6. Validasi status
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
