package pricing

import (
	"reflect"

	"github.com/tubes-cc/logistics/domain"
	"go.uber.org/mock/gomock"
)

// MockPricingRepository is a mock of PricingRepository interface.
type MockPricingRepository struct {
	ctrl     *gomock.Controller
	recorder *MockPricingRepositoryMockRecorder
}

// MockPricingRepositoryMockRecorder is the mock recorder for MockPricingRepository.
type MockPricingRepositoryMockRecorder struct {
	mock *MockPricingRepository
}

// NewMockPricingRepository creates a new mock instance.
func NewMockPricingRepository(ctrl *gomock.Controller) *MockPricingRepository {
	mock := &MockPricingRepository{ctrl: ctrl}
	mock.recorder = &MockPricingRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPricingRepository) EXPECT() *MockPricingRepositoryMockRecorder {
	return m.recorder
}

// GetBaseTariff mocks base method.
func (m *MockPricingRepository) GetBaseTariff(origin, destination string, serviceType domain.ServiceType) (float64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBaseTariff", origin, destination, serviceType)
	ret0, _ := ret[0].(float64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBaseTariff indicates an expected call of GetBaseTariff.
func (mr *MockPricingRepositoryMockRecorder) GetBaseTariff(origin, destination, serviceType interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBaseTariff", reflect.TypeOf((*MockPricingRepository)(nil).GetBaseTariff), origin, destination, serviceType)
}
