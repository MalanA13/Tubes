package order

import (
	"context"
	"errors"
	"testing"
	"time"

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

		res, err := service.CreateOrder(ctx, "user-1", req)
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
		res, err := service.CreateOrder(ctx, "", req)
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

		res, err := service.CreateOrder(ctx, "", req)
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

		res, err := service.CreateOrder(ctx, "", req)
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestOrderService_Ownership(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockOrderRepository(ctrl)
	mockPricing := NewMockPricingClient(ctrl)
	mockTracking := NewMockTrackingClient(ctrl)

	service := NewService(mockRepo, mockTracking, mockPricing)

	ctx := context.Background()

	t.Run("CreateOrder_PersistsUserID", func(t *testing.T) {
		req := domain.OrderRequest{
			SenderName:    "Alice",
			RecipientName: "Bob",
			Origin:        "Jakarta",
			Destination:   "Bandung",
			Weight:        1.0,
			ServiceType:   domain.ServiceRegular,
		}

		pricingRes := &domain.PricingResult{TotalCost: 10000.0}

		mockPricing.EXPECT().CalculatePrice(ctx, gomock.Any()).Return(pricingRes, nil)

		mockRepo.EXPECT().SaveOrder(gomock.Any()).DoAndReturn(func(order *OrderModel) error {
			assert.Equal(t, "user-42", order.UserID)
			return nil
		})

		mockTracking.EXPECT().AddTrackingEvent(ctx, gomock.Any()).Return(nil)

		res, err := service.CreateOrder(ctx, "user-42", req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("GetOrderByResiID_Success", func(t *testing.T) {
		mockRepo.EXPECT().GetOrderByResiID("AWB-001").Return(
			&OrderModel{ResiID: "AWB-001", UserID: "user-1"}, nil,
		)

		order, err := service.GetOrderByResiID("user-1", "AWB-001")
		assert.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, "AWB-001", order.ResiID)
	})

	t.Run("GetOrderByResiID_NotFound", func(t *testing.T) {
		mockRepo.EXPECT().GetOrderByResiID("AWB-999").Return(nil, nil)

		order, err := service.GetOrderByResiID("user-1", "AWB-999")
		assert.ErrorIs(t, err, domain.ErrShipmentNotFound)
		assert.Nil(t, order)
	})

	t.Run("GetOrderByResiID_ForbiddenAccess", func(t *testing.T) {
		mockRepo.EXPECT().GetOrderByResiID("AWB-001").Return(
			&OrderModel{ResiID: "AWB-001", UserID: "user-1"}, nil,
		)

		order, err := service.GetOrderByResiID("user-99", "AWB-001")
		assert.ErrorIs(t, err, domain.ErrForbidden)
		assert.Nil(t, order)
	})

	t.Run("ListOrdersByUserID_ReturnsList", func(t *testing.T) {
		mockRepo.EXPECT().ListOrdersByUserID("user-1").Return(
			[]*OrderModel{{OrderID: "ORD-1"}, {OrderID: "ORD-2"}}, nil,
		)

		result, err := service.ListOrdersByUserID("user-1")
		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("ListOrdersByUserID_EmptyList", func(t *testing.T) {
		mockRepo.EXPECT().ListOrdersByUserID("user-empty").Return(
			[]*OrderModel{}, nil,
		)

		result, err := service.ListOrdersByUserID("user-empty")
		assert.NoError(t, err)
		assert.Len(t, result, 0)
	})
}

func TestOrderService_GetCurrentStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockOrderRepository(ctrl)
	mockPricing := NewMockPricingClient(ctrl)
	mockTracking := NewMockTrackingClient(ctrl)

	service := NewService(mockRepo, mockTracking, mockPricing)
	ctx := context.Background()

	t.Run("TrackerReturnsStatus", func(t *testing.T) {
		mockTracking.EXPECT().GetCurrentStatus(ctx, "AWB-001").Return(domain.StatusDelivered, nil)
		result := service.GetCurrentStatus(ctx, "AWB-001", domain.StatusCreated)
		assert.Equal(t, domain.StatusDelivered, result)
	})

	t.Run("TrackerFails_ReturnsFallback", func(t *testing.T) {
		mockTracking.EXPECT().GetCurrentStatus(ctx, "AWB-002").Return(domain.TrackingStatus(""), errors.New("tracking unavailable"))
		result := service.GetCurrentStatus(ctx, "AWB-002", domain.StatusCreated)
		assert.Equal(t, domain.StatusCreated, result)
	})

	t.Run("TrackerReturnsCreated_WhenFreshResi", func(t *testing.T) {
		// Simulates the 404→CREATED path from HTTPTrackingClient
		mockTracking.EXPECT().GetCurrentStatus(ctx, "AWB-003").Return(domain.StatusCreated, nil)
		result := service.GetCurrentStatus(ctx, "AWB-003", domain.StatusCreated)
		assert.Equal(t, domain.StatusCreated, result)
	})
}

func TestOrderService_GetOrderTracking(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockOrderRepository(ctrl)
	mockPricing := NewMockPricingClient(ctrl)
	mockTracking := NewMockTrackingClient(ctrl)

	service := NewService(mockRepo, mockTracking, mockPricing)
	ctx := context.Background()

	t.Run("Success_WithEvents", func(t *testing.T) {
		// Arrange
		mockRepo.EXPECT().GetOrderByResiID("AWB-001").Return(
			&OrderModel{ResiID: "AWB-001", UserID: "user-1", Status: domain.StatusCreated}, nil,
		)
		mockTracking.EXPECT().GetCurrentStatus(ctx, "AWB-001").Return(domain.StatusInTransit, nil)
		mockTracking.EXPECT().GetTrackingHistory(ctx, "AWB-001").Return([]domain.TrackingEvent{
			{Status: domain.StatusCreated, Location: "Jakarta", Note: "Order Created", CreatedAt: time.Now()},
			{Status: domain.StatusInHub, Location: "HUB-JKT-01", Note: "Arrived at hub", CreatedAt: time.Now()},
			{Status: domain.StatusInTransit, Location: "HUB-JKT-01", Note: "Left hub", CreatedAt: time.Now()},
		}, nil)

		// Act
		result, err := service.GetOrderTracking(ctx, "user-1", "AWB-001")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "AWB-001", result.ResiID)
		assert.Equal(t, domain.StatusInTransit, result.CurrentStatus)
		assert.Len(t, result.Events, 3)
		assert.Equal(t, domain.StatusCreated, result.Events[0].Status)
		assert.Equal(t, "Jakarta", result.Events[0].Location)
	})

	t.Run("OrderNotFound", func(t *testing.T) {
		mockRepo.EXPECT().GetOrderByResiID("AWB-999").Return(nil, nil)

		result, err := service.GetOrderTracking(ctx, "user-1", "AWB-999")
		assert.ErrorIs(t, err, domain.ErrShipmentNotFound)
		assert.Nil(t, result)
	})

	t.Run("ForbiddenAccess", func(t *testing.T) {
		mockRepo.EXPECT().GetOrderByResiID("AWB-001").Return(
			&OrderModel{ResiID: "AWB-001", UserID: "user-1"}, nil,
		)

		result, err := service.GetOrderTracking(ctx, "user-99", "AWB-001")
		assert.ErrorIs(t, err, domain.ErrForbidden)
		assert.Nil(t, result)
	})

	t.Run("TrackingUnavailable_ReturnsFallback", func(t *testing.T) {
		mockRepo.EXPECT().GetOrderByResiID("AWB-002").Return(
			&OrderModel{ResiID: "AWB-002", UserID: "user-1", Status: domain.StatusCreated}, nil,
		)
		// GetCurrentStatus fails → fallback to OrderModel.Status
		mockTracking.EXPECT().GetCurrentStatus(ctx, "AWB-002").Return(domain.TrackingStatus(""), errors.New("tracking down"))
		// GetTrackingHistory fails → empty events
		mockTracking.EXPECT().GetTrackingHistory(ctx, "AWB-002").Return(nil, errors.New("tracking down"))

		result, err := service.GetOrderTracking(ctx, "user-1", "AWB-002")
		// Must succeed (no error) — graceful degradation
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.StatusCreated, result.CurrentStatus) // fallback to OrderModel.Status
		assert.Empty(t, result.Events)                              // empty slice, not nil
	})
}
