package functional

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tubes-cc/logistics/domain"
	"github.com/tubes-cc/logistics/internal/order"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// stubPricingClient always returns a fixed price.
type stubPricingClient struct{}

func (s *stubPricingClient) CalculatePrice(ctx context.Context, req domain.PricingRequest) (*domain.PricingResult, error) {
	return &domain.PricingResult{TotalCost: 50000.0, BaseCost: 10000.0}, nil
}

type simpleStubTrackingClient struct {
	events []*domain.TrackingEvent
}

func (s *simpleStubTrackingClient) AddTrackingEvent(ctx context.Context, e *domain.TrackingEvent) error {
	s.events = append(s.events, e)
	return nil
}

func (s *simpleStubTrackingClient) GetCurrentStatus(_ context.Context, _ string) (domain.TrackingStatus, error) {
	return domain.StatusCreated, nil
}

func (s *simpleStubTrackingClient) GetTrackingHistory(_ context.Context, _ string) ([]domain.TrackingEvent, error) {
	return []domain.TrackingEvent{}, nil
}

func TestOrderFunctional(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	repo, err := order.NewOrderRepository(db)
	require.NoError(t, err)

	pricingStub := &stubPricingClient{}
	trackingStub := &simpleStubTrackingClient{}

	service := order.NewService(repo, trackingStub, pricingStub)

	t.Run("Create Order Functional", func(t *testing.T) {
		req := domain.OrderRequest{
			SenderName:    "John Doe",
			RecipientName: "Jane Smith",
			Origin:        "Jakarta",
			Destination:   "Surabaya",
			Weight:        5.0,
			Dimensions:    "10x10x10",
			ItemType:      "Electronics",
			ServiceType:   domain.ServiceRegular,
			Distance:      800.0,
		}

		res, err := service.CreateOrder(context.Background(), "functional-test-user", req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.NotEmpty(t, res.OrderID)
		assert.NotEmpty(t, res.ResiID)
		assert.Equal(t, domain.StatusCreated, res.Status)
		assert.Equal(t, 50000.0, res.TotalCost)

		// Verify DB record
		var orderModel order.OrderModel
		err = db.Where("order_id = ?", res.OrderID).First(&orderModel).Error
		assert.NoError(t, err)
		assert.Equal(t, "John Doe", orderModel.SenderName)
		assert.Equal(t, 50000.0, orderModel.TotalCost)

		// Verify Tracking Event
		assert.Len(t, trackingStub.events, 1)
		assert.Equal(t, res.ResiID, trackingStub.events[0].ResiID)
		assert.Equal(t, domain.StatusCreated, trackingStub.events[0].Status)
	})
}
