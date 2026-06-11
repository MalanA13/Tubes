package handler

import (
	"encoding/json"
	"net/http"

	"github.com/tubes-cc/logistics/domain"
	"github.com/tubes-cc/logistics/internal/pricing"
	"github.com/tubes-cc/logistics/internal/response"
)

// HandleSendPricingHTTP handles pricing requests
func HandleSendPricingHTTP(service pricing.PricingService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req domain.PricingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, err.Error())
			return
		}

		res, err := service.CalculatePrice(req)
		if err != nil {
			response.InternalServerError(w, err.Error())
			return
		}

		response.OK(w, res)
	}
}
