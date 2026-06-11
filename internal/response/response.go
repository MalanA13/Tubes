package response

import (
	"encoding/json"
	"net/http"
)

// ErrorDetail represents the standard error object inside the response.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// APIResponse represents the standard API response structure.
type APIResponse struct {
	Success bool         `json:"success"`
	Data    interface{}  `json:"data,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
}

// writeJSON is a helper to write the JSON response.
func writeJSON(w http.ResponseWriter, status int, resp APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// OK sends a 200 OK response with data.
func OK(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

// Created sends a 201 Created response with data.
func Created(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    data,
	})
}

// BadRequest sends a 400 error response.
func BadRequest(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusBadRequest, APIResponse{
		Success: false,
		Error: &ErrorDetail{
			Code:    "BAD_REQUEST",
			Message: message,
		},
	})
}

// Unauthorized sends a 401 error response.
func Unauthorized(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusUnauthorized, APIResponse{
		Success: false,
		Error: &ErrorDetail{
			Code:    "UNAUTHORIZED",
			Message: message,
		},
	})
}

// Forbidden sends a 403 error response.
func Forbidden(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusForbidden, APIResponse{
		Success: false,
		Error: &ErrorDetail{
			Code:    "FORBIDDEN",
			Message: message,
		},
	})
}

// NotFound sends a 404 error response.
func NotFound(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusNotFound, APIResponse{
		Success: false,
		Error: &ErrorDetail{
			Code:    "NOT_FOUND",
			Message: message,
		},
	})
}

// MethodNotAllowed sends a 405 error response.
func MethodNotAllowed(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusMethodNotAllowed, APIResponse{
		Success: false,
		Error: &ErrorDetail{
			Code:    "METHOD_NOT_ALLOWED",
			Message: message,
		},
	})
}

// InternalServerError sends a 500 error response.
func InternalServerError(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusInternalServerError, APIResponse{
		Success: false,
		Error: &ErrorDetail{
			Code:    "INTERNAL_SERVER_ERROR",
			Message: message,
		},
	})
}
