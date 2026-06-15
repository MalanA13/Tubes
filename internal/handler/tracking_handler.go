package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tubes-cc/logistics/domain"
	"github.com/tubes-cc/logistics/internal/response"
	"github.com/tubes-cc/logistics/internal/tracking"
)

type TrackingRequest struct {
	ResiID   string                `json:"resi_id"`
	Status   domain.TrackingStatus `json:"status"`
	Location string                `json:"location"`
	Note     string                `json:"note"`
}

// HandleSendTrackingHTTP handles tracking event addition requests
func HandleSendTrackingHTTP(service tracking.TrackingService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req TrackingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, err.Error())
			return
		}

		err := service.AddTrackingEvent(req.ResiID, req.Status, req.Location, req.Note)
		if err != nil {
			response.InternalServerError(w, err.Error())
			return
		}

		response.Created(w, map[string]string{"message": "tracking event added"})
	}
}

// HandleGetTrackingHTTP retrieves the tracking history of a given resiID.
func HandleGetTrackingHTTP(service tracking.TrackingService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		resiID := vars["resiID"]
		if resiID == "" {
			response.BadRequest(w, "resiID is required")
			return
		}

		events, err := service.GetHistory(resiID)
		if err != nil {
			response.InternalServerError(w, err.Error())
			return
		}

		if len(events) == 0 {
			response.NotFound(w, "tracking history not found for this resi")
			return
		}

		response.OK(w, events)
	}
}

// HandleGetCurrentStatusHTTP returns the current shipment status snapshot.
func HandleGetCurrentStatusHTTP(service tracking.TrackingService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		resiID := vars["resiID"]
		if resiID == "" {
			response.BadRequest(w, "resiID is required")
			return
		}

		status, err := service.GetCurrentStatus(resiID)
		if err != nil {
			if errors.Is(err, domain.ErrShipmentNotFound) {
				response.NotFound(w, "shipment status not found")
				return
			}
			response.InternalServerError(w, err.Error())
			return
		}

		response.OK(w, map[string]string{
			"resi_id": resiID,
			"status":  string(status),
		})
	}
}
