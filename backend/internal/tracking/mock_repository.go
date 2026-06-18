package tracking

import (
	"reflect"

	"github.com/tubes-cc/logistics/domain"
	"go.uber.org/mock/gomock"
)

// MockTrackingRepository is a mock of TrackingRepository interface.
type MockTrackingRepository struct {
	ctrl     *gomock.Controller
	recorder *MockTrackingRepositoryMockRecorder
}

// MockTrackingRepositoryMockRecorder is the mock recorder for MockTrackingRepository.
type MockTrackingRepositoryMockRecorder struct {
	mock *MockTrackingRepository
}

// NewMockTrackingRepository creates a new mock instance.
func NewMockTrackingRepository(ctrl *gomock.Controller) *MockTrackingRepository {
	mock := &MockTrackingRepository{ctrl: ctrl}
	mock.recorder = &MockTrackingRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTrackingRepository) EXPECT() *MockTrackingRepositoryMockRecorder {
	return m.recorder
}

// GetTrackingHistory mocks base method.
func (m *MockTrackingRepository) GetTrackingHistory(resiID string) ([]domain.TrackingEvent, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTrackingHistory", resiID)
	ret0, _ := ret[0].([]domain.TrackingEvent)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTrackingHistory indicates an expected call of GetTrackingHistory.
func (mr *MockTrackingRepositoryMockRecorder) GetTrackingHistory(resiID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTrackingHistory", reflect.TypeOf((*MockTrackingRepository)(nil).GetTrackingHistory), resiID)
}

// SaveEvent mocks base method.
func (m *MockTrackingRepository) SaveEvent(event domain.TrackingEvent) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveEvent", event)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveEvent indicates an expected call of SaveEvent.
func (mr *MockTrackingRepositoryMockRecorder) SaveEvent(event interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveEvent", reflect.TypeOf((*MockTrackingRepository)(nil).SaveEvent), event)
}

// UpdateShipmentStatus mocks base method.
func (m *MockTrackingRepository) UpdateShipmentStatus(resiID string, status domain.TrackingStatus) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateShipmentStatus", resiID, status)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateShipmentStatus indicates an expected call of UpdateShipmentStatus.
func (mr *MockTrackingRepositoryMockRecorder) UpdateShipmentStatus(resiID, status interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateShipmentStatus", reflect.TypeOf((*MockTrackingRepository)(nil).UpdateShipmentStatus), resiID, status)
}

// GetCurrentStatus mocks base method.
func (m *MockTrackingRepository) GetCurrentStatus(resiID string) (domain.TrackingStatus, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCurrentStatus", resiID)
	ret0, _ := ret[0].(domain.TrackingStatus)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCurrentStatus indicates an expected call of GetCurrentStatus.
func (mr *MockTrackingRepositoryMockRecorder) GetCurrentStatus(resiID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCurrentStatus", reflect.TypeOf((*MockTrackingRepository)(nil).GetCurrentStatus), resiID)
}

// ListAllShipments mocks base method.
func (m *MockTrackingRepository) ListAllShipments() ([]ShipmentModel, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAllShipments")
	ret0, _ := ret[0].([]ShipmentModel)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAllShipments indicates an expected call of ListAllShipments.
func (mr *MockTrackingRepositoryMockRecorder) ListAllShipments() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAllShipments", reflect.TypeOf((*MockTrackingRepository)(nil).ListAllShipments))
}
