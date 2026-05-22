package order

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tubes-cc/logistics/domain"
	"go.uber.org/mock/gomock"
)

func TestOrderService_CreateOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockOrderRepository(ctrl)
	mockPricing := NewMockPricingClient(ctrl)
	mockTracking := NewMockTrackingClient(ctrl)

	service := NewService(mockRepo, mockTracking, mockPricing)

	ctx := context.Background()

	t.Run("Successfully Create Order", func(t *testing.T) {
		req := domain.OrderRequest{
			SenderName:    "Alice",
			RecipientName: "Bob",
			Origin:        "Jakarta",
			Destination:   "Bandung",
			Weight:        2.0,
			ServiceType:   domain.ServiceRegular,
		}

		pricingRes := &domain.PricingResult{
			TotalCost: 20000.0,
		}

		// Expect Pricing Calculation
		mockPricing.EXPECT().CalculatePrice(ctx, gomock.Any()).Return(pricingRes, nil)

		// Expect Save Order
		mockRepo.EXPECT().SaveOrder(gomock.Any()).DoAndReturn(func(order *OrderModel) error {
			assert.Equal(t, "Alice", order.SenderName)
			assert.Equal(t, "Bob", order.RecipientName)
			assert.Equal(t, 20000.0, order.TotalCost)
			assert.NotEmpty(t, order.OrderID)
			assert.NotEmpty(t, order.ResiID)
			return nil
		})

		// Expect Tracking Event
		mockTracking.EXPECT().AddTrackingEvent(ctx, gomock.Any()).Return(nil)

		res, err := service.CreateOrder(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, 20000.0, res.TotalCost)
		assert.Equal(t, domain.StatusCreated, res.Status)
		assert.NotEmpty(t, res.OrderID)
		assert.NotEmpty(t, res.ResiID)
	})

	t.Run("Missing Sender Name", func(t *testing.T) {
		req := domain.OrderRequest{
			RecipientName: "Bob",
		}
		res, err := service.CreateOrder(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "sender and recipient names are required", err.Error())
	})

	t.Run("Pricing Client Error", func(t *testing.T) {
		req := domain.OrderRequest{
			SenderName:    "Alice",
			RecipientName: "Bob",
		}
		mockPricing.EXPECT().CalculatePrice(ctx, gomock.Any()).Return(nil, errors.New("pricing error"))

		res, err := service.CreateOrder(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Contains(t, err.Error(), "pricing error")
	})

	t.Run("Repository Save Error", func(t *testing.T) {
		req := domain.OrderRequest{
			SenderName:    "Alice",
			RecipientName: "Bob",
		}
		pricingRes := &domain.PricingResult{TotalCost: 15000.0}

		mockPricing.EXPECT().CalculatePrice(ctx, gomock.Any()).Return(pricingRes, nil)
		mockRepo.EXPECT().SaveOrder(gomock.Any()).Return(errors.New("db error"))

		res, err := service.CreateOrder(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Contains(t, err.Error(), "db error")
	})
}
