// Package handler — HTTP handler untuk Hub Service.
// Menjadi adapter antara HTTP layer dan service layer.
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/tubes-cc/logistics/hub"
)

// HubHandler menangani semua HTTP request untuk Hub Service.
type HubHandler struct {
	service *hub.Service
}

// NewHubHandler membuat handler baru.
func NewHubHandler(service *hub.Service) *HubHandler {
	return &HubHandler{service: service}
}

// scanRequest adalah body request untuk scan-in dan scan-out.
type scanRequest struct {
	ResiID string `json:"resi_id"`
	HubID  string `json:"hub_id"`
}

// ScanIn menangani POST /hub/scan-in
//
// Request body:
//
//	{"resi_id": "JNE-001", "hub_id": "HUB-JKT-01"}
//
// Response 200:
//
//	{"status": "ok", "message": "Scan-in berhasil"}
func (h *HubHandler) ScanIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "gunakan POST")
		return
	}

	var req scanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "body request tidak valid JSON")
		return
	}

	if err := h.service.ScanIn(r.Context(), req.ResiID, req.HubID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, "Scan-in berhasil dicatat")
}

// ScanOut menangani POST /hub/scan-out
//
// Request body:
//
//	{"resi_id": "JNE-001", "hub_id": "HUB-JKT-01"}
func (h *HubHandler) ScanOut(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "gunakan POST")
		return
	}

	var req scanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "body request tidak valid JSON")
		return
	}

	if err := h.service.ScanOut(r.Context(), req.ResiID, req.HubID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, "Scan-out berhasil dicatat")
}

// ---- shared response helpers ----

type successResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func writeSuccess(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(successResponse{Status: "ok", Message: msg})
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(errorResponse{Error: http.StatusText(code), Message: msg})
}
