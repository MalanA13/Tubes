// Package courier — mock untuk semua interface yang digunakan Courier Service.
// Berisi mock: Repository, TrackingClient, OrderClient.
package courier

import (
	"context"
	"reflect"

	"github.com/tubes-cc/logistics/domain"
	"go.uber.org/mock/gomock"
)

// ================================================================
// MockRepository
// ================================================================

type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
}

type MockRepositoryMockRecorder struct{ mock *MockRepository }

func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	mock := &MockRepository{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder{mock}
	return mock
}

func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder { return m.recorder }

func (m *MockRepository) GetShipment(ctx context.Context, resiID string) (*domain.Shipment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetShipment", ctx, resiID)
	ret0, _ := ret[0].(*domain.Shipment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockRepositoryMockRecorder) GetShipment(ctx, resiID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetShipment",
		reflect.TypeOf((*MockRepository)(nil).GetShipment), ctx, resiID)
}

func (m *MockRepository) UpdateShipment(ctx context.Context, shipment *domain.Shipment) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateShipment", ctx, shipment)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockRepositoryMockRecorder) UpdateShipment(ctx, shipment interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateShipment",
		reflect.TypeOf((*MockRepository)(nil).UpdateShipment), ctx, shipment)
}

func (m *MockRepository) CreateShipmentIfNotExists(ctx context.Context, resiID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateShipmentIfNotExists", ctx, resiID)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockRepositoryMockRecorder) CreateShipmentIfNotExists(ctx, resiID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateShipmentIfNotExists",
		reflect.TypeOf((*MockRepository)(nil).CreateShipmentIfNotExists), ctx, resiID)
}

func (m *MockRepository) ListShipmentsByCourier(ctx context.Context, courierUserID string) ([]*domain.Shipment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListShipmentsByCourier", ctx, courierUserID)
	ret0, _ := ret[0].([]*domain.Shipment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockRepositoryMockRecorder) ListShipmentsByCourier(ctx, courierUserID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListShipmentsByCourier",
		reflect.TypeOf((*MockRepository)(nil).ListShipmentsByCourier), ctx, courierUserID)
}

// ================================================================
// MockTrackingClient
// ================================================================

type MockTrackingClient struct {
	ctrl     *gomock.Controller
	recorder *MockTrackingClientMockRecorder
}

type MockTrackingClientMockRecorder struct{ mock *MockTrackingClient }

func NewMockTrackingClient(ctrl *gomock.Controller) *MockTrackingClient {
	mock := &MockTrackingClient{ctrl: ctrl}
	mock.recorder = &MockTrackingClientMockRecorder{mock}
	return mock
}

func (m *MockTrackingClient) EXPECT() *MockTrackingClientMockRecorder { return m.recorder }

func (m *MockTrackingClient) AddTrackingEvent(ctx context.Context, event *domain.TrackingEvent) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddTrackingEvent", ctx, event)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockTrackingClientMockRecorder) AddTrackingEvent(ctx, event interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddTrackingEvent",
		reflect.TypeOf((*MockTrackingClient)(nil).AddTrackingEvent), ctx, event)
}

func (m *MockTrackingClient) GetCurrentStatus(ctx context.Context, resiID string) (domain.TrackingStatus, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCurrentStatus", ctx, resiID)
	ret0, _ := ret[0].(domain.TrackingStatus)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockTrackingClientMockRecorder) GetCurrentStatus(ctx, resiID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCurrentStatus",
		reflect.TypeOf((*MockTrackingClient)(nil).GetCurrentStatus), ctx, resiID)
}

func (m *MockTrackingClient) GetTrackingHistory(ctx context.Context, resiID string) ([]domain.TrackingEvent, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTrackingHistory", ctx, resiID)
	ret0, _ := ret[0].([]domain.TrackingEvent)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockTrackingClientMockRecorder) GetTrackingHistory(ctx, resiID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTrackingHistory",
		reflect.TypeOf((*MockTrackingClient)(nil).GetTrackingHistory), ctx, resiID)
}

// ================================================================
// MockOrderClient
// ================================================================

type MockOrderClient struct {
	ctrl     *gomock.Controller
	recorder *MockOrderClientMockRecorder
}

type MockOrderClientMockRecorder struct{ mock *MockOrderClient }

func NewMockOrderClient(ctrl *gomock.Controller) *MockOrderClient {
	mock := &MockOrderClient{ctrl: ctrl}
	mock.recorder = &MockOrderClientMockRecorder{mock}
	return mock
}

func (m *MockOrderClient) EXPECT() *MockOrderClientMockRecorder { return m.recorder }

func (m *MockOrderClient) ValidateResi(ctx context.Context, resiID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateResi", ctx, resiID)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockOrderClientMockRecorder) ValidateResi(ctx, resiID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateResi",
		reflect.TypeOf((*MockOrderClient)(nil).ValidateResi), ctx, resiID)
}

// ================================================================
// MockAuthClient
// ================================================================

type MockAuthClient struct {
	ctrl     *gomock.Controller
	recorder *MockAuthClientMockRecorder
}

type MockAuthClientMockRecorder struct{ mock *MockAuthClient }

func NewMockAuthClient(ctrl *gomock.Controller) *MockAuthClient {
	mock := &MockAuthClient{ctrl: ctrl}
	mock.recorder = &MockAuthClientMockRecorder{mock}
	return mock
}

func (m *MockAuthClient) EXPECT() *MockAuthClientMockRecorder { return m.recorder }

func (m *MockAuthClient) ValidateToken(ctx context.Context, token string) (*domain.AuthClaims, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateToken", ctx, token)
	ret0, _ := ret[0].(*domain.AuthClaims)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockAuthClientMockRecorder) ValidateToken(ctx, token interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateToken",
		reflect.TypeOf((*MockAuthClient)(nil).ValidateToken), ctx, token)
}

func (m *MockAuthClient) ValidateUserRole(ctx context.Context, userID string, expectedRole domain.UserRole) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateUserRole", ctx, userID, expectedRole)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockAuthClientMockRecorder) ValidateUserRole(ctx, userID, expectedRole interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateUserRole",
		reflect.TypeOf((*MockAuthClient)(nil).ValidateUserRole), ctx, userID, expectedRole)
}
