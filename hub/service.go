// Package hub — business logic untuk Hub Service.
// Bergantung pada: Repository (interface), client.TrackingClient, client.OrderClient.
// Menggunakan UUID untuk setiap tracking event ID agar unik.
package hub

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tubes-cc/logistics/client"
	"github.com/tubes-cc/logistics/domain"
)

// Service menyediakan operasi bisnis Hub Service.
// Semua dependency di-inject melalui constructor.
type Service struct {
	repo            Repository           // akses data lokal
	trackingClient  client.TrackingClient // kirim event ke Tracking Service
	orderClient     client.OrderClient    // validasi resi ke Order Service
}

// NewService membuat instance Hub Service baru.
//
// Semua parameter adalah interface sehingga mudah di-mock saat testing.
//
// Contoh:
//   svc := hub.NewService(sqliteRepo, httpTrackingClient, httpOrderClient)
func NewService(repo Repository, tc client.TrackingClient, oc client.OrderClient) *Service {
	return &Service{
		repo:           repo,
		trackingClient: tc,
		orderClient:    oc,
	}
}

// ScanIn mencatat paket tiba (masuk) di hub.
//
// Flow:
//  1. Validasi resiID dan hubID tidak kosong
//  2. Validasi resi ke Order Service
//  3. Pastikan shipment ada di DB (buat jika belum ada)
//  4. Cek tidak ada duplicate scan-in di hub yang sama
//  5. Update status → IN_HUB
//  6. Kirim tracking event ke Tracking Service
func (s *Service) ScanIn(ctx context.Context, resiID, hubID string) error {
	// 1. Validasi input
	if resiID == "" {
		return domain.ErrInvalidResiID
	}
	if hubID == "" {
		return domain.ErrInvalidHubID
	}

	// 2. Validasi resi ke Order Service
	if err := s.orderClient.ValidateResi(ctx, resiID); err != nil {
		return fmt.Errorf("scan-in: validasi resi gagal: %w", err)
	}

	// 3. Inisialisasi shipment di DB jika belum ada
	if err := s.repo.CreateShipmentIfNotExists(ctx, resiID); err != nil {
		return fmt.Errorf("scan-in: gagal inisialisasi shipment: %w", err)
	}

	// 4. Ambil data shipment & cek duplicate
	shipment, err := s.repo.GetShipment(ctx, resiID)
	if err != nil {
		return fmt.Errorf("scan-in: gagal ambil shipment: %w", err)
	}

	if shipment.Status == domain.StatusInHub && shipment.HubID == hubID {
		return domain.ErrAlreadyScannedIn
	}

	// 5. Update status
	now := time.Now()
	shipment.Status = domain.StatusInHub
	shipment.HubID = hubID
	shipment.UpdatedAt = now

	if err := s.repo.UpdateShipment(ctx, shipment); err != nil {
		return fmt.Errorf("scan-in: gagal update shipment: %w", err)
	}

	// 6. Kirim tracking event — ID menggunakan UUID agar selalu unik
	event := &domain.TrackingEvent{
		ID:        uuid.New().String(),
		ResiID:    resiID,
		Status:    domain.StatusInHub,
		Location:  hubID,
		Note:      fmt.Sprintf("Paket tiba di hub %s", hubID),
		CreatedAt: now,
	}

	if err := s.trackingClient.AddTrackingEvent(ctx, event); err != nil {
		// Non-fatal: log dan lanjutkan (jangan gagalkan operasi utama)
		return fmt.Errorf("scan-in: gagal kirim tracking event: %w", err)
	}

	return nil
}

// ScanOut mencatat paket keluar dari hub.
//
// Flow:
//  1. Validasi input
//  2. Validasi resi ke Order Service
//  3. Ambil shipment — status harus IN_HUB
//  4. Update status → IN_TRANSIT
//  5. Kirim tracking event
func (s *Service) ScanOut(ctx context.Context, resiID, hubID string) error {
	// 1. Validasi input
	if resiID == "" {
		return domain.ErrInvalidResiID
	}
	if hubID == "" {
		return domain.ErrInvalidHubID
	}

	// 2. Validasi resi
	if err := s.orderClient.ValidateResi(ctx, resiID); err != nil {
		return fmt.Errorf("scan-out: validasi resi gagal: %w", err)
	}

	// 3. Ambil & validasi status
	shipment, err := s.repo.GetShipment(ctx, resiID)
	if err != nil {
		return fmt.Errorf("scan-out: gagal ambil shipment: %w", err)
	}

	if shipment.Status != domain.StatusInHub {
		return fmt.Errorf("scan-out: status saat ini '%s', harus 'IN_HUB': %w",
			shipment.Status, domain.ErrInvalidStatus)
	}

	// 4. Update status
	now := time.Now()
	shipment.Status = domain.StatusInTransit
	shipment.HubID = hubID
	shipment.UpdatedAt = now

	if err := s.repo.UpdateShipment(ctx, shipment); err != nil {
		return fmt.Errorf("scan-out: gagal update shipment: %w", err)
	}

	// 5. Kirim tracking event dengan UUID
	event := &domain.TrackingEvent{
		ID:        uuid.New().String(),
		ResiID:    resiID,
		Status:    domain.StatusInTransit,
		Location:  hubID,
		Note:      fmt.Sprintf("Paket keluar dari hub %s", hubID),
		CreatedAt: now,
	}

	if err := s.trackingClient.AddTrackingEvent(ctx, event); err != nil {
		return fmt.Errorf("scan-out: gagal kirim tracking event: %w", err)
	}

	return nil
}
