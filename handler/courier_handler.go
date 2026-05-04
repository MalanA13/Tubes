// Package handler — HTTP handler untuk Courier Service.
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/tubes-cc/logistics/courier"
	"github.com/tubes-cc/logistics/domain"
)

// CourierHandler menangani semua HTTP request untuk Courier Service.
type CourierHandler struct {
	service *courier.Service
}

// NewCourierHandler membuat handler baru.
func NewCourierHandler(service *courier.Service) *CourierHandler {
	return &CourierHandler{service: service}
}

// assignRequest adalah body request untuk assign courier.
type assignRequest struct {
	ResiID    string `json:"resi_id"`
	CourierID string `json:"courier_id"`
}

// deliveryStatusRequest adalah body request untuk update status pengiriman.
type deliveryStatusRequest struct {
	ResiID   string                `json:"resi_id"`
	Status   domain.TrackingStatus `json:"status"`
	ProofURL string                `json:"proof_url"`
}

// AssignCourier menangani POST /courier/assign
//
// Request body:
//
//	{"resi_id": "JNE-001", "courier_id": "KURIR-BUDI-01"}
//
// Response 200:
//
//	{"status": "ok", "message": "Kurir berhasil ditugaskan"}
func (h *CourierHandler) AssignCourier(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "gunakan POST")
		return
	}

	var req assignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "body request tidak valid JSON")
		return
	}

	if err := h.service.AssignCourier(r.Context(), req.ResiID, req.CourierID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, "Kurir berhasil ditugaskan")
}

// UpdateDeliveryStatus menangani POST /courier/delivery-status
//
// Request body:
//
//	{"resi_id": "JNE-001", "status": "DELIVERED", "proof_url": "https://..."}
//
// Status yang valid: DELIVERED (wajib proof_url), FAILED, RETURNED
func (h *CourierHandler) UpdateDeliveryStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "gunakan POST")
		return
	}

	var req deliveryStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "body request tidak valid JSON")
		return
	}

	if err := h.service.UpdateDeliveryStatus(r.Context(), req.ResiID, req.Status, req.ProofURL); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, "Status pengiriman berhasil diperbarui")
}
