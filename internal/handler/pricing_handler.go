package handler

import (
	"encoding/json"
	"net/http"

	"github.com/tubes-cc/logistics/domain"
	"github.com/tubes-cc/logistics/internal/pricing"
)

// HandleSendPricingHTTP handles pricing requests
func HandleSendPricingHTTP(service pricing.PricingService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req domain.PricingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res, err := service.CalculatePrice(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	}
}
