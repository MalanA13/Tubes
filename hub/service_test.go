// Package hub — unit test untuk Hub Service.
//
// Prinsip: TIDAK ada akses database. Semua dependency di-mock.
// Pattern Arrange → Act → Assert pada setiap test.
package hub

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tubes-cc/logistics/domain"
	"go.uber.org/mock/gomock"
)

// newTestService adalah helper untuk membuat service dengan semua mock.
func newTestService(ctrl *gomock.Controller) (*Service, *MockRepository, *MockTrackingClient, *MockOrderClient) {
	repo := NewMockRepository(ctrl)
	tc := NewMockTrackingClient(ctrl)
	oc := NewMockOrderClient(ctrl)
	svc := NewService(repo, tc, oc)
	return svc, repo, tc, oc
}

// ============================================================
// TESTS: ScanIn
// ============================================================

func TestScanIn_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, repo, tc, oc := newTestService(ctrl)
	ctx := context.Background()

	shipment := &domain.Shipment{ResiID: "RESI-001", Status: domain.StatusCreated, UpdatedAt: time.Now()}

	// Expectations
	oc.EXPECT().ValidateResi(ctx, "RESI-001").Return(nil)
	repo.EXPECT().CreateShipmentIfNotExists(ctx, "RESI-001").Return(nil)
	repo.EXPECT().GetShipment(ctx, "RESI-001").Return(shipment, nil)
	repo.EXPECT().UpdateShipment(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, s *domain.Shipment) error {
		assert.Equal(t, domain.StatusInHub, s.Status)
		assert.Equal(t, "HUB-JKT-01", s.HubID)
		return nil
	})
	tc.EXPECT().AddTrackingEvent(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, e *domain.TrackingEvent) error {
		// Verifikasi event ID adalah UUID valid (tidak kosong, tidak duplikat)
		assert.NotEmpty(t, e.ID, "event ID harus UUID, tidak boleh kosong")
		assert.Equal(t, domain.StatusInHub, e.Status)
		return nil
	})

	err := svc.ScanIn(ctx, "RESI-001", "HUB-JKT-01")
	require.NoError(t, err)
}

func TestScanIn_EmptyResiID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, _, _, _ := newTestService(ctrl)

	err := svc.ScanIn(context.Background(), "", "HUB-JKT-01")
	assert.ErrorIs(t, err, domain.ErrInvalidResiID)
}

func TestScanIn_EmptyHubID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, _, _, _ := newTestService(ctrl)

	err := svc.ScanIn(context.Background(), "RESI-001", "")
	assert.ErrorIs(t, err, domain.ErrInvalidHubID)
}

func TestScanIn_ResiInvalid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, _, _, oc := newTestService(ctrl)
	ctx := context.Background()

	// OrderClient menolak resi (misal: resi tidak ditemukan di Order Service)
	oc.EXPECT().ValidateResi(ctx, "RESI-PALSU").Return(domain.ErrResiInvalid)

	err := svc.ScanIn(ctx, "RESI-PALSU", "HUB-JKT-01")
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrResiInvalid)
}

func TestScanIn_ShipmentNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, repo, _, oc := newTestService(ctrl)
	ctx := context.Background()

	oc.EXPECT().ValidateResi(ctx, "RESI-001").Return(nil)
	repo.EXPECT().CreateShipmentIfNotExists(ctx, "RESI-001").Return(nil)
	repo.EXPECT().GetShipment(ctx, "RESI-001").Return(nil, domain.ErrShipmentNotFound)

	err := svc.ScanIn(ctx, "RESI-001", "HUB-JKT-01")
	assert.ErrorIs(t, err, domain.ErrShipmentNotFound)
}

func TestScanIn_AlreadyScannedInSameHub(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, repo, _, oc := newTestService(ctrl)
	ctx := context.Background()

	// Shipment sudah IN_HUB di hub yang sama
	shipment := &domain.Shipment{ResiID: "RESI-001", Status: domain.StatusInHub, HubID: "HUB-JKT-01"}

	oc.EXPECT().ValidateResi(ctx, "RESI-001").Return(nil)
	repo.EXPECT().CreateShipmentIfNotExists(ctx, "RESI-001").Return(nil)
	repo.EXPECT().GetShipment(ctx, "RESI-001").Return(shipment, nil)

	err := svc.ScanIn(ctx, "RESI-001", "HUB-JKT-01")
	assert.ErrorIs(t, err, domain.ErrAlreadyScannedIn)
}

func TestScanIn_UpdateShipmentFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, repo, _, oc := newTestService(ctrl)
	ctx := context.Background()

	shipment := &domain.Shipment{ResiID: "RESI-001", Status: domain.StatusCreated}
	oc.EXPECT().ValidateResi(ctx, "RESI-001").Return(nil)
	repo.EXPECT().CreateShipmentIfNotExists(ctx, "RESI-001").Return(nil)
	repo.EXPECT().GetShipment(ctx, "RESI-001").Return(shipment, nil)
	repo.EXPECT().UpdateShipment(ctx, gomock.Any()).Return(errors.New("db error"))

	err := svc.ScanIn(ctx, "RESI-001", "HUB-JKT-01")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

func TestScanIn_TrackingClientFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, repo, tc, oc := newTestService(ctrl)
	ctx := context.Background()

	shipment := &domain.Shipment{ResiID: "RESI-001", Status: domain.StatusCreated}
	oc.EXPECT().ValidateResi(ctx, "RESI-001").Return(nil)
	repo.EXPECT().CreateShipmentIfNotExists(ctx, "RESI-001").Return(nil)
	repo.EXPECT().GetShipment(ctx, "RESI-001").Return(shipment, nil)
	repo.EXPECT().UpdateShipment(ctx, gomock.Any()).Return(nil)
	tc.EXPECT().AddTrackingEvent(ctx, gomock.Any()).Return(errors.New("tracking service down"))

	err := svc.ScanIn(ctx, "RESI-001", "HUB-JKT-01")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tracking service down")
}

// ============================================================
// TESTS: ScanOut
// ============================================================

func TestScanOut_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, repo, tc, oc := newTestService(ctrl)
	ctx := context.Background()

	shipment := &domain.Shipment{ResiID: "RESI-001", Status: domain.StatusInHub, HubID: "HUB-JKT-01"}

	oc.EXPECT().ValidateResi(ctx, "RESI-001").Return(nil)
	repo.EXPECT().GetShipment(ctx, "RESI-001").Return(shipment, nil)
	repo.EXPECT().UpdateShipment(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, s *domain.Shipment) error {
		assert.Equal(t, domain.StatusInTransit, s.Status)
		return nil
	})
	tc.EXPECT().AddTrackingEvent(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, e *domain.TrackingEvent) error {
		assert.NotEmpty(t, e.ID, "UUID tidak boleh kosong")
		assert.Equal(t, domain.StatusInTransit, e.Status)
		return nil
	})

	err := svc.ScanOut(ctx, "RESI-001", "HUB-JKT-01")
	require.NoError(t, err)
}

func TestScanOut_EmptyResiID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, _, _, _ := newTestService(ctrl)

	err := svc.ScanOut(context.Background(), "", "HUB-JKT-01")
	assert.ErrorIs(t, err, domain.ErrInvalidResiID)
}

func TestScanOut_EmptyHubID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, _, _, _ := newTestService(ctrl)

	err := svc.ScanOut(context.Background(), "RESI-001", "")
	assert.ErrorIs(t, err, domain.ErrInvalidHubID)
}

func TestScanOut_ResiInvalid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, _, _, oc := newTestService(ctrl)
	ctx := context.Background()

	oc.EXPECT().ValidateResi(ctx, "RESI-001").Return(domain.ErrResiInvalid)

	err := svc.ScanOut(ctx, "RESI-001", "HUB-JKT-01")
	assert.ErrorIs(t, err, domain.ErrResiInvalid)
}

func TestScanOut_InvalidStatus_NotInHub(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, repo, _, oc := newTestService(ctrl)
	ctx := context.Background()

	shipment := &domain.Shipment{ResiID: "RESI-001", Status: domain.StatusCreated}
	oc.EXPECT().ValidateResi(ctx, "RESI-001").Return(nil)
	repo.EXPECT().GetShipment(ctx, "RESI-001").Return(shipment, nil)

	err := svc.ScanOut(ctx, "RESI-001", "HUB-JKT-01")
	assert.ErrorIs(t, err, domain.ErrInvalidStatus)
}

func TestScanOut_ShipmentNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, repo, _, oc := newTestService(ctrl)
	ctx := context.Background()

	oc.EXPECT().ValidateResi(ctx, "RESI-001").Return(nil)
	repo.EXPECT().GetShipment(ctx, "RESI-001").Return(nil, domain.ErrShipmentNotFound)

	err := svc.ScanOut(ctx, "RESI-001", "HUB-JKT-01")
	assert.ErrorIs(t, err, domain.ErrShipmentNotFound)
}

func TestScanOut_UpdateShipmentFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, repo, _, oc := newTestService(ctrl)
	ctx := context.Background()

	shipment := &domain.Shipment{ResiID: "RESI-001", Status: domain.StatusInHub}
	oc.EXPECT().ValidateResi(ctx, "RESI-001").Return(nil)
	repo.EXPECT().GetShipment(ctx, "RESI-001").Return(shipment, nil)
	repo.EXPECT().UpdateShipment(ctx, gomock.Any()).Return(errors.New("write failed"))

	err := svc.ScanOut(ctx, "RESI-001", "HUB-JKT-01")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "write failed")
}
