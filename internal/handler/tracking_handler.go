package handler

import (
	"encoding/json"
	"net/http"

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
