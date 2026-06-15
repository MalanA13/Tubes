package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/tubes-cc/logistics/domain"
	"github.com/tubes-cc/logistics/internal/contextutil"
	"github.com/tubes-cc/logistics/internal/order"
	"github.com/tubes-cc/logistics/internal/response"
)

// orderDetailResponse is the response shape for order detail and list endpoints.
type orderDetailResponse struct {
	OrderID       string                `json:"order_id"`
	ResiID        string                `json:"resi_id"`
	SenderName    string                `json:"sender_name"`
	RecipientName string                `json:"recipient_name"`
	Origin        string                `json:"origin"`
	Destination   string                `json:"destination"`
	Weight        float64               `json:"weight"`
	ItemType      string                `json:"item_type"`
	ServiceType   domain.ServiceType    `json:"service_type"`
	TotalCost     float64               `json:"total_cost"`
	Status        domain.TrackingStatus `json:"status"`
	CreatedAt     time.Time             `json:"created_at"`
}

// toOrderDetailResponse maps an OrderModel to an orderDetailResponse.
func toOrderDetailResponse(m *order.OrderModel) orderDetailResponse {
	return orderDetailResponse{
		OrderID:       m.OrderID,
		ResiID:        m.ResiID,
		SenderName:    m.SenderName,
		RecipientName: m.RecipientName,
		Origin:        m.Origin,
		Destination:   m.Destination,
		Weight:        m.Weight,
		ItemType:      m.ItemType,
		ServiceType:   m.ServiceType,
		TotalCost:     m.TotalCost,
		Status:        m.Status,
		CreatedAt:     m.CreatedAt,
	}
}

// toOrderDetailResponseWithStatus maps an OrderModel to an orderDetailResponse,
// overriding the status with the provided live status from Tracking Service.
func toOrderDetailResponseWithStatus(m *order.OrderModel, status domain.TrackingStatus) orderDetailResponse {
	return orderDetailResponse{
		OrderID:       m.OrderID,
		ResiID:        m.ResiID,
		SenderName:    m.SenderName,
		RecipientName: m.RecipientName,
		Origin:        m.Origin,
		Destination:   m.Destination,
		Weight:        m.Weight,
		ItemType:      m.ItemType,
		ServiceType:   m.ServiceType,
		TotalCost:     m.TotalCost,
		Status:        status, // live status from Tracking, not orderModel.Status
		CreatedAt:     m.CreatedAt,
	}
}

// HandleOrderHTTP handles order creation requests.
func HandleOrderHTTP(service order.OrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(contextutil.AuthClaimsKey).(*domain.AuthClaims)
		if !ok || claims == nil {
			response.Unauthorized(w, "authentication required")
			return
		}

		var req domain.OrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, err.Error())
			return
		}

		res, err := service.CreateOrder(r.Context(), claims.UserID, req)
		if err != nil {
			response.InternalServerError(w, err.Error())
			return
		}

		response.Created(w, res)
	}
}

// HandleGetOrderHTTP handles retrieving a single order by resiID.
func HandleGetOrderHTTP(service order.OrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resiID := mux.Vars(r)["resiID"]
		if resiID == "" {
			response.BadRequest(w, "resiID is required")
			return
		}

		claims, ok := r.Context().Value(contextutil.AuthClaimsKey).(*domain.AuthClaims)
		if !ok || claims == nil {
			response.Unauthorized(w, "authentication required")
			return
		}

		orderModel, err := service.GetOrderByResiID(claims.UserID, resiID)
		if err != nil {
			if errors.Is(err, domain.ErrShipmentNotFound) {
				response.NotFound(w, "order not found")
				return
			}
			if errors.Is(err, domain.ErrForbidden) {
				response.Forbidden(w, "access denied")
				return
			}
			response.InternalServerError(w, err.Error())
			return
		}

		liveStatus := service.GetCurrentStatus(r.Context(), resiID, orderModel.Status)
		response.OK(w, toOrderDetailResponseWithStatus(orderModel, liveStatus))
	}
}

// HandleListOrdersHTTP handles listing all orders for the authenticated user.
func HandleListOrdersHTTP(service order.OrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(contextutil.AuthClaimsKey).(*domain.AuthClaims)
		if !ok || claims == nil {
			response.Unauthorized(w, "authentication required")
			return
		}

		orders, err := service.ListOrdersByUserID(claims.UserID)
		if err != nil {
			response.InternalServerError(w, err.Error())
			return
		}

		result := make([]orderDetailResponse, 0, len(orders))
		for _, o := range orders {
			liveStatus := service.GetCurrentStatus(r.Context(), o.ResiID, o.Status)
			result = append(result, toOrderDetailResponseWithStatus(o, liveStatus))
		}

		response.OK(w, result)
	}
}

// HandleGetOrderTrackingHTTP handles GET /orders/{resiID}/tracking.
// Returns the customer-facing shipment timeline including current status and event history.
// Requires authentication. Enforces order ownership (403 for foreign orders, 404 for missing).
func HandleGetOrderTrackingHTTP(service order.OrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resiID := mux.Vars(r)["resiID"]
		if resiID == "" {
			response.BadRequest(w, "resiID is required")
			return
		}

		claims, ok := r.Context().Value(contextutil.AuthClaimsKey).(*domain.AuthClaims)
		if !ok || claims == nil {
			response.Unauthorized(w, "authentication required")
			return
		}

		tracking, err := service.GetOrderTracking(r.Context(), claims.UserID, resiID)
		if err != nil {
			if errors.Is(err, domain.ErrShipmentNotFound) {
				response.NotFound(w, "order not found")
				return
			}
			if errors.Is(err, domain.ErrForbidden) {
				response.Forbidden(w, "access denied")
				return
			}
			response.InternalServerError(w, err.Error())
			return
		}

		response.OK(w, tracking)
	}
}
