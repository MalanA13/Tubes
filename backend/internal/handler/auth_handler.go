package handler

import (
	"encoding/json"
	"net/http"

	models "github.com/tubes-cc/logistics/domain"
	"github.com/tubes-cc/logistics/internal/auth"
	"github.com/tubes-cc/logistics/internal/response"
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
			response.MethodNotAllowed(w, "method not allowed")
			return
		}

		var req models.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, "invalid request body")
			return
		}

		token, err := s.Login(req.Email, req.Password)
		if err != nil {
			response.Unauthorized(w, err.Error())
			return
		}

		response.OK(w, models.LoginResponse{Token: token})
	}
}

func HandleRegister(s auth.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.MethodNotAllowed(w, "method not allowed")
			return
		}

		var req models.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, "invalid request body")
			return
		}

		user, err := s.Register(req.FullName, req.Email, req.Password)
		if err != nil {
			response.BadRequest(w, err.Error())
			return
		}

		response.Created(w, models.RegisterResponse{
			ID:        user.ID,
			Email:     user.Email,
			FullName:  user.FullName,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
		})
	}
}

func HandleValidateToken(s auth.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.MethodNotAllowed(w, "method not allowed")
			return
		}

		var req validateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, "invalid request body")
			return
		}

		claims, err := s.ValidateToken(req.Token)
		if err != nil {
			response.Unauthorized(w, "invalid token: "+err.Error())
			return
		}

		response.OK(w, validateResponse{
			UserID: claims.UserID,
			Role:   string(claims.Role),
		})
	}
}

type validateUserRoleRequest struct {
	UserID       string `json:"user_id"`
	ExpectedRole string `json:"expected_role"`
}

func HandleValidateUserRole(s auth.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.MethodNotAllowed(w, "method not allowed")
			return
		}

		var req validateUserRoleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, "invalid request body")
			return
		}

		if req.UserID == "" || req.ExpectedRole == "" {
			response.BadRequest(w, "user_id and expected_role are required")
			return
		}

		// Convert string role to domain.UserRole
		expectedRole := models.UserRole(req.ExpectedRole)

		if err := s.ValidateUserRole(r.Context(), req.UserID, expectedRole); err != nil {
			response.Forbidden(w, err.Error())
			return
		}

		response.OK(w, map[string]interface{}{
			"valid": true,
			"message": "user has the expected role",
		})
	}
}
