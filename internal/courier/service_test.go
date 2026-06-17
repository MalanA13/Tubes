// Package courier — unit test untuk Courier Service.
//
// Prinsip: TIDAK ada akses database. Semua dependency di-mock.
package courier

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

// newTestService adalah helper membuat service dengan semua mock.
func newTestService(ctrl *gomock.Controller) (*Service, *MockRepository, *MockTrackingClient, *MockOrderClient, *MockAuthClient) {
	repo := NewMockRepository(ctrl)
	tc := NewMockTrackingClient(ctrl)
	oc := NewMockOrderClient(ctrl)
	ac := NewMockAuthClient(ctrl)
	svc := NewService(repo, tc, oc, ac)
	return svc, repo, tc, oc, ac
}

// ============================================================
// TESTS: AssignCourier
// ============================================================

func TestAssignCourier_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, repo, tc, oc, ac := newTestService(ctrl)
	ctx := context.Background()

	shipment := &domain.Shipment{
		ResiID: "RESI-001", Status: domain.StatusInTransit,
		HubID: "HUB-JKT-01", CourierID: "", UpdatedAt: time.Now(),
	}

	oc.EXPECT().ValidateResi(ctx, "RESI-001").Return(nil)
	repo.EXPECT().CreateShipmentIfNotExists(ctx, "RESI-001").Return(nil)
	ac.EXPECT().ValidateUserRole(ctx, "KURIR-001", domain.RoleCourier).Return(nil)
	repo.EXPECT().GetShipment(ctx, "RESI-001").Return(shipment, nil)
	repo.EXPECT().UpdateShipment(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, s *domain.Shipment) error {
		assert.Equal(t, domain.StatusOutDelivery, s.Status)
		assert.Equal(t, "KURIR-001", s.CourierID)
		return nil
	})
	tc.EXPECT().AddTrackingEvent(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, e *domain.TrackingEvent) error {
		assert.NotEmpty(t, e.ID, "UUID tidak boleh kosong")
		assert.Equal(t, domain.StatusOutDelivery, e.Status)
		return nil
	})

	err := svc.AssignCourier(ctx, "RESI-001", "KURIR-001")
	require.NoError(t, err)
}

func TestAssignCourier_EmptyResiID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, _, _, _, _ := newTestService(ctrl)

	err := svc.AssignCourier(context.Background(), "", "KURIR-001")
	assert.ErrorIs(t, err, domain.ErrInvalidResiID)
}

func TestAssignCourier_EmptyCourierID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, _, _, _, _ := newTestService(ctrl)

	err := svc.AssignCourier(context.Background(), "RESI-001", "")
	assert.ErrorIs(t, err, domain.ErrInvalidCourierID)
}

func TestAssignCourier_ResiInvalid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, _, _, oc, ac := newTestService(ctrl)
	ctx := context.Background()

	ac.EXPECT().ValidateUserRole(ctx, "KURIR-001", domain.RoleCourier).Return(nil)
	oc.EXPECT().ValidateResi(ctx, "RESI-001").Return(domain.ErrResiInvalid)

	err := svc.AssignCourier(ctx, "RESI-001", "KURIR-001")
	assert.ErrorIs(t, err, domain.ErrResiInvalid)
}

func TestAssignCourier_ShipmentNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, repo, _, oc, ac := newTestService(ctrl)
	ctx := context.Background()

	oc.EXPECT().ValidateResi(ctx, "RESI-001").Return(nil)
	repo.EXPECT().CreateShipmentIfNotExists(ctx, "RESI-001").Return(nil)
	ac.EXPECT().ValidateUserRole(ctx, "KURIR-001", domain.RoleCourier).Return(nil)
	repo.EXPECT().GetShipment(ctx, "RESI-001").Return(nil, domain.ErrShipmentNotFound)

	err := svc.AssignCourier(ctx, "RESI-001", "KURIR-001")
	assert.ErrorIs(t, err, domain.ErrShipmentNotFound)
}

func TestAssignCourier_InvalidStatus_NotInTransit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, repo, _, oc, ac := newTestService(ctrl)
	ctx := context.Background()

	// Masih IN_HUB, belum scan-out
	shipment := &domain.Shipment{ResiID: "RESI-001", Status: domain.StatusInHub}
	oc.EXPECT().ValidateResi(ctx, "RESI-001").Return(nil)
	repo.EXPECT().CreateShipmentIfNotExists(ctx, "RESI-001").Return(nil)
	ac.EXPECT().ValidateUserRole(ctx, "KURIR-001", domain.RoleCourier).Return(nil)
	repo.EXPECT().GetShipment(ctx, "RESI-001").Return(shipment, nil)

	err := svc.AssignCourier(ctx, "RESI-001", "KURIR-001")
	assert.ErrorIs(t, err, domain.ErrInvalidStatus)
}

func TestAssignCourier_AlreadyAssigned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, repo, _, oc, ac := newTestService(ctrl)
	ctx := context.Background()

	shipment := &domain.Shipment{
		ResiID: "RESI-001", Status: domain.StatusInTransit, CourierID: "KURIR-LAMA",
	}
	oc.EXPECT().ValidateResi(ctx, "RESI-001").Return(nil)
	repo.EXPECT().CreateShipmentIfNotExists(ctx, "RESI-001").Return(nil)
	ac.EXPECT().ValidateUserRole(ctx, "KURIR-BARU", domain.RoleCourier).Return(nil)
	repo.EXPECT().GetShipment(ctx, "RESI-001").Return(shipment, nil)

	err := svc.AssignCourier(ctx, "RESI-001", "KURIR-BARU")
	assert.ErrorIs(t, err, domain.ErrCourierAlreadyAssigned)
}

func TestAssignCourier_TrackingClientFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, repo, tc, oc, ac := newTestService(ctrl)
	ctx := context.Background()

	shipment := &domain.Shipment{ResiID: "RESI-001", Status: domain.StatusInTransit}
	oc.EXPECT().ValidateResi(ctx, "RESI-001").Return(nil)
	repo.EXPECT().CreateShipmentIfNotExists(ctx, "RESI-001").Return(nil)
	ac.EXPECT().ValidateUserRole(ctx, "KURIR-001", domain.RoleCourier).Return(nil)
	repo.EXPECT().GetShipment(ctx, "RESI-001").Return(shipment, nil)
	repo.EXPECT().UpdateShipment(ctx, gomock.Any()).Return(nil)
	tc.EXPECT().AddTrackingEvent(ctx, gomock.Any()).Return(errors.New("network timeout"))

	err := svc.AssignCourier(ctx, "RESI-001", "KURIR-001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "network timeout")
}

// ============================================================
// TESTS: UpdateDeliveryStatus
// ============================================================

func TestUpdateDeliveryStatus_Success_Delivered(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, repo, tc, oc, _ := newTestService(ctrl)
	ctx := context.Background()

	claims := &domain.AuthClaims{UserID: "KURIR-001", Role: domain.RoleCourier}
	shipment := &domain.Shipment{
		ResiID: "RESI-001", Status: domain.StatusOutDelivery, CourierID: "KURIR-001",
		CourierUserID: "KURIR-001",
	}
	oc.EXPECT().ValidateResi(ctx, "RESI-001").Return(nil)
	repo.EXPECT().GetShipment(ctx, "RESI-001").Return(shipment, nil)
	repo.EXPECT().UpdateShipment(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, s *domain.Shipment) error {
		assert.Equal(t, domain.StatusDelivered, s.Status)
		assert.Equal(t, "https://proof.example.com/foto.jpg", s.ProofURL)
		return nil
	})
	tc.EXPECT().AddTrackingEvent(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, e *domain.TrackingEvent) error {
		assert.NotEmpty(t, e.ID)
		assert.Equal(t, domain.StatusDelivered, e.Status)
		return nil
	})

	err := svc.UpdateDeliveryStatus(ctx, claims, "RESI-001", domain.StatusDelivered, "https://proof.example.com/foto.jpg")
	require.NoError(t, err)
}

func TestUpdateDeliveryStatus_Success_Failed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, repo, tc, oc, _ := newTestService(ctrl)
	ctx := context.Background()

	claims := &domain.AuthClaims{UserID: "KURIR-001", Role: domain.RoleCourier}
	shipment := &domain.Shipment{
		ResiID: "RESI-001", Status: domain.StatusOutDelivery, CourierID: "KURIR-001",
		CourierUserID: "KURIR-001",
	}
	oc.EXPECT().ValidateResi(ctx, "RESI-001").Return(nil)
	repo.EXPECT().GetShipment(ctx, "RESI-001").Return(shipment, nil)
	repo.EXPECT().UpdateShipment(ctx, gomock.Any()).Return(nil)
	tc.EXPECT().AddTrackingEvent(ctx, gomock.Any()).Return(nil)

	// FAILED tidak memerlukan proofURL
	err := svc.UpdateDeliveryStatus(ctx, claims, "RESI-001", domain.StatusFailed, "")
	require.NoError(t, err)
}

func TestUpdateDeliveryStatus_Success_Returned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, repo, tc, oc, _ := newTestService(ctrl)
	ctx := context.Background()

	claims := &domain.AuthClaims{UserID: "KURIR-001", Role: domain.RoleCourier}
	shipment := &domain.Shipment{ResiID: "RESI-001", Status: domain.StatusOutDelivery, CourierID: "KURIR-001", CourierUserID: "KURIR-001"}
	oc.EXPECT().ValidateResi(ctx, "RESI-001").Return(nil)
	repo.EXPECT().GetShipment(ctx, "RESI-001").Return(shipment, nil)
	repo.EXPECT().UpdateShipment(ctx, gomock.Any()).Return(nil)
	tc.EXPECT().AddTrackingEvent(ctx, gomock.Any()).Return(nil)

	err := svc.UpdateDeliveryStatus(ctx, claims, "RESI-001", domain.StatusReturned, "")
	require.NoError(t, err)
}

func TestUpdateDeliveryStatus_EmptyResiID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, _, _, _, _ := newTestService(ctrl)

	claims := &domain.AuthClaims{UserID: "KURIR-001", Role: domain.RoleCourier}
	err := svc.UpdateDeliveryStatus(context.Background(), claims, "", domain.StatusDelivered, "proof.jpg")
	assert.ErrorIs(t, err, domain.ErrInvalidResiID)
}

func TestUpdateDeliveryStatus_InvalidStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, _, _, _, _ := newTestService(ctrl)

	claims := &domain.AuthClaims{UserID: "KURIR-001", Role: domain.RoleCourier}
	// IN_HUB bukan status delivery yang valid
	err := svc.UpdateDeliveryStatus(context.Background(), claims, "RESI-001", domain.StatusInHub, "")
	assert.ErrorIs(t, err, domain.ErrInvalidStatus)
}

func TestUpdateDeliveryStatus_DeliveredWithoutProof(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, _, _, _, _ := newTestService(ctrl)

	claims := &domain.AuthClaims{UserID: "KURIR-001", Role: domain.RoleCourier}
	// DELIVERED wajib ada proofURL
	err := svc.UpdateDeliveryStatus(context.Background(), claims, "RESI-001", domain.StatusDelivered, "")
	assert.ErrorIs(t, err, domain.ErrInvalidProofURL)
}

func TestUpdateDeliveryStatus_ResiInvalid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, _, _, oc, _ := newTestService(ctrl)
	ctx := context.Background()

	claims := &domain.AuthClaims{UserID: "KURIR-001", Role: domain.RoleCourier}
	oc.EXPECT().ValidateResi(ctx, "RESI-001").Return(domain.ErrResiInvalid)

	err := svc.UpdateDeliveryStatus(ctx, claims, "RESI-001", domain.StatusFailed, "")
	assert.ErrorIs(t, err, domain.ErrResiInvalid)
}

func TestUpdateDeliveryStatus_InvalidCurrentStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, repo, _, oc, _ := newTestService(ctrl)
	ctx := context.Background()

	claims := &domain.AuthClaims{UserID: "KURIR-001", Role: domain.RoleCourier}
	// Status IN_HUB — bukan OUT_DELIVERY
	shipment := &domain.Shipment{ResiID: "RESI-001", Status: domain.StatusInHub, CourierUserID: "KURIR-001"}
	oc.EXPECT().ValidateResi(ctx, "RESI-001").Return(nil)
	repo.EXPECT().GetShipment(ctx, "RESI-001").Return(shipment, nil)

	err := svc.UpdateDeliveryStatus(ctx, claims, "RESI-001", domain.StatusDelivered, "proof.jpg")
	assert.ErrorIs(t, err, domain.ErrInvalidStatus)
}

func TestUpdateDeliveryStatus_UpdateFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc, repo, _, oc, _ := newTestService(ctrl)
	ctx := context.Background()

	claims := &domain.AuthClaims{UserID: "KURIR-001", Role: domain.RoleCourier}
	shipment := &domain.Shipment{ResiID: "RESI-001", Status: domain.StatusOutDelivery, CourierID: "KURIR-001", CourierUserID: "KURIR-001"}
	oc.EXPECT().ValidateResi(ctx, "RESI-001").Return(nil)
	repo.EXPECT().GetShipment(ctx, "RESI-001").Return(shipment, nil)
	repo.EXPECT().UpdateShipment(ctx, gomock.Any()).Return(errors.New("db write failed"))

	err := svc.UpdateDeliveryStatus(ctx, claims, "RESI-001", domain.StatusFailed, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db write failed")
}
