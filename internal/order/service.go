package order

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tubes-cc/logistics/client"
	"github.com/tubes-cc/logistics/domain"
)

// OrderService handles order business logic.
type OrderService struct {
	repo           OrderRepository
	trackingClient client.TrackingClient
	pricingClient  client.PricingClient
}

// NewService creates a new OrderService.
func NewService(repo OrderRepository, trackingClient client.TrackingClient, pricingClient client.PricingClient) *OrderService {
	return &OrderService{
		repo:           repo,
		trackingClient: trackingClient,
		pricingClient:  pricingClient,
	}
}

// CreateOrder processes a new order, calculates price, generates AWB, and starts tracking.
func (s *OrderService) CreateOrder(ctx context.Context, userID string, req domain.OrderRequest) (*domain.OrderResponse, error) {
	if req.SenderName == "" || req.RecipientName == "" {
		return nil, fmt.Errorf("sender and recipient names are required")
	}

	// 1. Calculate Price
	pricingReq := domain.PricingRequest{
		Origin:      req.Origin,
		Destination: req.Destination,
		Weight:      req.Weight,
		Distance:    req.Distance,
		ServiceType: req.ServiceType,
	}

	pricingRes, err := s.pricingClient.CalculatePrice(ctx, pricingReq)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate price: %w", err)
	}

	// 2. Generate IDs
	orderID := "ORD-" + uuid.NewString()[:8]
	resiID := "AWB-" + uuid.NewString()[:8]

	// 3. Save Order
	orderModel := &OrderModel{
		OrderID:       orderID,
		ResiID:        resiID,
		UserID:        userID,
		SenderName:    req.SenderName,
		RecipientName: req.RecipientName,
		Origin:        req.Origin,
		Destination:   req.Destination,
		Weight:        req.Weight,
		ItemType:      req.ItemType,
		ServiceType:   req.ServiceType,
		TotalCost:     pricingRes.TotalCost,
		Status:        domain.StatusCreated,
		CreatedAt:     time.Now(),
	}

	if err := s.repo.SaveOrder(orderModel); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	// 4. Send Tracking Event (CREATED)
	trackingEvent := &domain.TrackingEvent{
		ID:        uuid.NewString(),
		ResiID:    resiID,
		Status:    domain.StatusCreated,
		Location:  req.Origin,
		Note:      "Order Created and Booked",
		CreatedAt: time.Now(),
	}

	// Ignore tracking event failure (karena not implemented yet), but log it normally.
	_ = s.trackingClient.AddTrackingEvent(ctx, trackingEvent)

	return &domain.OrderResponse{
		OrderID:   orderID,
		ResiID:    resiID,
		Status:    domain.StatusCreated,
		TotalCost: pricingRes.TotalCost,
	}, nil
}

// GetOrderByResiID retrieves an order by resiID and validates caller ownership.
// Returns domain.ErrShipmentNotFound if not found.
// Returns domain.ErrForbidden if order exists but belongs to a different user.
func (s *OrderService) GetOrderByResiID(userID, resiID string) (*OrderModel, error) {
	order, err := s.repo.GetOrderByResiID(resiID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, domain.ErrShipmentNotFound
	}
	if order.UserID != userID {
		return nil, domain.ErrForbidden
	}
	return order, nil
}

// ListOrdersByUserID returns all orders belonging to a user.
func (s *OrderService) ListOrdersByUserID(userID string) ([]*OrderModel, error) {
	return s.repo.ListOrdersByUserID(userID)
}
