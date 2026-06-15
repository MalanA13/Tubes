package order

import (
	"context"
	"reflect"

	"github.com/tubes-cc/logistics/domain"
	"go.uber.org/mock/gomock"
)

// MockOrderRepository is a mock of OrderRepository interface.
type MockOrderRepository struct {
	ctrl     *gomock.Controller
	recorder *MockOrderRepositoryMockRecorder
}
type MockOrderRepositoryMockRecorder struct {
	mock *MockOrderRepository
}

func NewMockOrderRepository(ctrl *gomock.Controller) *MockOrderRepository {
	mock := &MockOrderRepository{ctrl: ctrl}
	mock.recorder = &MockOrderRepositoryMockRecorder{mock}
	return mock
}
func (m *MockOrderRepository) EXPECT() *MockOrderRepositoryMockRecorder {
	return m.recorder
}
func (m *MockOrderRepository) SaveOrder(order *OrderModel) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveOrder", order)
	ret0, _ := ret[0].(error)
	return ret0
}
func (mr *MockOrderRepositoryMockRecorder) SaveOrder(order interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveOrder", reflect.TypeOf((*MockOrderRepository)(nil).SaveOrder), order)
}

func (m *MockOrderRepository) ValidateResi(resiID string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateResi", resiID)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}
func (mr *MockOrderRepositoryMockRecorder) ValidateResi(resiID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateResi", reflect.TypeOf((*MockOrderRepository)(nil).ValidateResi), resiID)
}

func (m *MockOrderRepository) GetOrderByResiID(resiID string) (*OrderModel, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrderByResiID", resiID)
	ret0, _ := ret[0].(*OrderModel)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}
func (mr *MockOrderRepositoryMockRecorder) GetOrderByResiID(resiID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrderByResiID", reflect.TypeOf((*MockOrderRepository)(nil).GetOrderByResiID), resiID)
}

func (m *MockOrderRepository) ListOrdersByUserID(userID string) ([]*OrderModel, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListOrdersByUserID", userID)
	ret0, _ := ret[0].([]*OrderModel)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}
func (mr *MockOrderRepositoryMockRecorder) ListOrdersByUserID(userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListOrdersByUserID", reflect.TypeOf((*MockOrderRepository)(nil).ListOrdersByUserID), userID)
}

// MockPricingClient is a mock of client.PricingClient interface.
type MockPricingClient struct {
	ctrl     *gomock.Controller
	recorder *MockPricingClientMockRecorder
}
type MockPricingClientMockRecorder struct {
	mock *MockPricingClient
}

func NewMockPricingClient(ctrl *gomock.Controller) *MockPricingClient {
	mock := &MockPricingClient{ctrl: ctrl}
	mock.recorder = &MockPricingClientMockRecorder{mock}
	return mock
}
func (m *MockPricingClient) EXPECT() *MockPricingClientMockRecorder {
	return m.recorder
}
func (m *MockPricingClient) CalculatePrice(ctx context.Context, req domain.PricingRequest) (*domain.PricingResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CalculatePrice", ctx, req)
	ret0, _ := ret[0].(*domain.PricingResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}
func (mr *MockPricingClientMockRecorder) CalculatePrice(ctx, req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CalculatePrice", reflect.TypeOf((*MockPricingClient)(nil).CalculatePrice), ctx, req)
}

// MockTrackingClient is a mock of client.TrackingClient interface.
type MockTrackingClient struct {
	ctrl     *gomock.Controller
	recorder *MockTrackingClientMockRecorder
}
type MockTrackingClientMockRecorder struct {
	mock *MockTrackingClient
}

func NewMockTrackingClient(ctrl *gomock.Controller) *MockTrackingClient {
	mock := &MockTrackingClient{ctrl: ctrl}
	mock.recorder = &MockTrackingClientMockRecorder{mock}
	return mock
}
func (m *MockTrackingClient) EXPECT() *MockTrackingClientMockRecorder {
	return m.recorder
}
func (m *MockTrackingClient) AddTrackingEvent(ctx context.Context, event *domain.TrackingEvent) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddTrackingEvent", ctx, event)
	ret0, _ := ret[0].(error)
	return ret0
}
func (mr *MockTrackingClientMockRecorder) AddTrackingEvent(ctx, event interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddTrackingEvent", reflect.TypeOf((*MockTrackingClient)(nil).AddTrackingEvent), ctx, event)
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
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCurrentStatus", reflect.TypeOf((*MockTrackingClient)(nil).GetCurrentStatus), ctx, resiID)
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
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTrackingHistory", reflect.TypeOf((*MockTrackingClient)(nil).GetTrackingHistory), ctx, resiID)
}
