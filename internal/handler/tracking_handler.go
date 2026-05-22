package handler

import (
	"encoding/json"
	"net/http"

	"github.com/tubes-cc/logistics/domain"
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
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := service.AddTrackingEvent(req.ResiID, req.Status, req.Location, req.Note)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message": "tracking event added"}`))
	}
}
