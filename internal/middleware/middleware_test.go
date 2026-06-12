package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tubes-cc/logistics/internal/contextutil"
	"github.com/tubes-cc/logistics/internal/middleware"
)

func TestRequestIDMiddleware_GeneratesNewID(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	var ctxReqID string
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxReqID = contextutil.GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.RequestIDMiddleware(nextHandler)
	handler.ServeHTTP(rr, req)

	// Check Context
	if ctxReqID == "" {
		t.Errorf("Expected RequestID in context, got empty string")
	}

	// Check Header
	headerReqID := rr.Header().Get(contextutil.RequestIDHeader)
	if headerReqID == "" {
		t.Errorf("Expected RequestID in response header, got empty string")
	}

	if ctxReqID != headerReqID {
		t.Errorf("Expected Context ID %s to match Header ID %s", ctxReqID, headerReqID)
	}
}

func TestRequestIDMiddleware_PreservesExistingID(t *testing.T) {
	existingID := "existing-uuid-123"
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set(contextutil.RequestIDHeader, existingID)
	rr := httptest.NewRecorder()

	var ctxReqID string
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxReqID = contextutil.GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.RequestIDMiddleware(nextHandler)
	handler.ServeHTTP(rr, req)

	if ctxReqID != existingID {
		t.Errorf("Expected context RequestID to be %s, got %s", existingID, ctxReqID)
	}

	headerReqID := rr.Header().Get(contextutil.RequestIDHeader)
	if headerReqID != existingID {
		t.Errorf("Expected header RequestID to be %s, got %s", existingID, headerReqID)
	}
}

func TestLoggingMiddleware_PreservesStatusCode(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	expectedStatus := http.StatusCreated
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(expectedStatus)
	})

	handler := middleware.LoggingMiddleware(nextHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != expectedStatus {
		t.Errorf("Expected status code %d, got %d", expectedStatus, rr.Code)
	}
}

func TestLoggingMiddleware_PreservesResponseBody(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	expectedBody := `{"status":"ok"}`
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(expectedBody))
	})

	handler := middleware.LoggingMiddleware(nextHandler)
	handler.ServeHTTP(rr, req)

	if rr.Body.String() != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, rr.Body.String())
	}
}

func TestMiddlewareChain_DoesNotAlterAPIResponse(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	type apiResponse struct {
		Message string `json:"message"`
	}
	expectedResponse := apiResponse{Message: "Success"}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(expectedResponse)
	})

	// Chain: RequestID -> Logging -> Handler
	handler := middleware.RequestIDMiddleware(middleware.LoggingMiddleware(nextHandler))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("Expected status code %d, got %d", http.StatusAccepted, rr.Code)
	}

	var actualResponse apiResponse
	err := json.Unmarshal(rr.Body.Bytes(), &actualResponse)
	if err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	if actualResponse.Message != expectedResponse.Message {
		t.Errorf("Expected message %s, got %s", expectedResponse.Message, actualResponse.Message)
	}

	if rr.Header().Get(contextutil.RequestIDHeader) == "" {
		t.Errorf("Expected RequestID header to be set by chain")
	}
}
