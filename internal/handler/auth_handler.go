package handler

import (
	"encoding/json"
	"net/http"

	models "github.com/tubes-cc/logistics/domain"
	"github.com/tubes-cc/logistics/internal/auth"
)

type validateRequest struct {
	Token string `json:"token"`
}

type validateResponse struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

func HandleLogin(s auth.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req models.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		token, err := s.Login(req.Email, req.Password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.LoginResponse{Token: token})
	}
}

func HandleValidateToken(s auth.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req validateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Token != "TOKEN_JWT_DUMMY" {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(validateResponse{
			UserID: "1",
			Role:   string(models.RoleAdmin),
		})
	}
}
