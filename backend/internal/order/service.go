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

// TrackingEventSummary is the customer-facing shape of a single tracking event.
// Drops internal fields (id, resi_id) and renames created_at to timestamp.
type TrackingEventSummary struct {
	Status    domain.TrackingStatus `json:"status"`
	Location  string                `json:"location"`
	Note      string                `json:"note"`
	Timestamp time.Time             `json:"timestamp"`
}

// OrderTracking is the combined tracking response returned to the customer.
type OrderTracking struct {
	ResiID        string                 `json:"resi_id"`
	CurrentStatus domain.TrackingStatus  `json:"current_status"`
	Events        []TrackingEventSummary `json:"events"`
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

// GetCurrentStatus fetches the current shipment status from Tracking Service.
// On any error (network, timeout, not found), returns the provided fallback status.
// This method never returns an error — the caller always gets a usable status.
func (s *OrderService) GetCurrentStatus(ctx context.Context, resiID string, fallback domain.TrackingStatus) domain.TrackingStatus {
	status, err := s.trackingClient.GetCurrentStatus(ctx, resiID)
	if err != nil {
		return fallback
	}
	return status
}

// GetOrderTracking retrieves the tracking timeline for an order.
// Validates that userID owns the order before returning data.
// Returns domain.ErrShipmentNotFound if no order exists for resiID.
// Returns domain.ErrForbidden if the order belongs to a different user.
// Never fails due to Tracking Service unavailability:
// current_status falls back to OrderModel.Status, events falls back to empty slice.
func (s *OrderService) GetOrderTracking(ctx context.Context, userID, resiID string) (*OrderTracking, error) {
	// 1. Verify ownership — reuse existing method
	orderModel, err := s.GetOrderByResiID(userID, resiID)
	if err != nil {
		return nil, err // propagate ErrShipmentNotFound or ErrForbidden as-is
	}

	// 2. Current status from Tracking (with graceful fallback)
	currentStatus := s.GetCurrentStatus(ctx, resiID, orderModel.Status)

	// 3. Event history from Tracking (degrade gracefully on error)
	rawEvents, trackingErr := s.trackingClient.GetTrackingHistory(ctx, resiID)
	if trackingErr != nil {
		rawEvents = []domain.TrackingEvent{}
	}

	// 4. Map to customer-facing shape — drop id and resi_id
	events := make([]TrackingEventSummary, 0, len(rawEvents))
	for _, e := range rawEvents {
		events = append(events, TrackingEventSummary{
			Status:    e.Status,
			Location:  e.Location,
			Note:      e.Note,
			Timestamp: e.CreatedAt,
		})
	}

	return &OrderTracking{
		ResiID:        resiID,
		CurrentStatus: currentStatus,
		Events:        events,
	}, nil
}

// ListAllOrders returns all orders in the system for admin use.
// No user filtering is applied.
func (s *OrderService) ListAllOrders(ctx context.Context) ([]*OrderModel, error) {
	orders, err := s.repo.ListAllOrders()
	if err != nil {
		return nil, fmt.Errorf("list-all-orders: %w", err)
	}
	return orders, nil
}
