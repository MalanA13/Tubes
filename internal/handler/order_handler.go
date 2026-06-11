package handler

import (
	"encoding/json"
	"net/http"

	"github.com/tubes-cc/logistics/domain"
	"github.com/tubes-cc/logistics/internal/order"
	"github.com/tubes-cc/logistics/internal/response"
)

// HandleOrderHTTP handles order creation requests.
func HandleOrderHTTP(service order.OrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req domain.OrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, err.Error())
			return
		}

		res, err := service.CreateOrder(r.Context(), req)
		if err != nil {
			response.InternalServerError(w, err.Error())
			return
		}

		response.Created(w, res)
	}
}
