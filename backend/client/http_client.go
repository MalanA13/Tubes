// Package client — implementasi HTTP konkret dari client interface.
// Di production, file ini digunakan untuk komunikasi nyata ke microservice lain.
// Di test, interface ini di-mock sehingga file ini tidak diperlukan.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/tubes-cc/logistics/domain"
	"github.com/tubes-cc/logistics/internal/contextutil"
)

// ================================================================
// HTTPTrackingClient — implementasi TrackingClient via HTTP
// ================================================================

// HTTPTrackingClient mengirim tracking event via HTTP POST ke Tracking Service.
type HTTPTrackingClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewHTTPTrackingClient membuat TrackingClient baru yang berkomunikasi via HTTP.
func NewHTTPTrackingClient(baseURL string) *HTTPTrackingClient {
	return &HTTPTrackingClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// AddTrackingEvent mengirim POST request ke Tracking Service.
func (c *HTTPTrackingClient) AddTrackingEvent(ctx context.Context, event *domain.TrackingEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("tracking client: marshal event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/events", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("tracking client: buat request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if reqID := contextutil.GetRequestID(ctx); reqID != "" {
		req.Header.Set(contextutil.RequestIDHeader, reqID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("tracking client: kirim request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("tracking client: status tidak OK: %d", resp.StatusCode)
	}

	return nil
}

// GetCurrentStatus mengambil status paket terkini dari Tracking Service via GET.
// Jika paket belum punya snapshot (404), dikembalikan domain.StatusCreated.
func (c *HTTPTrackingClient) GetCurrentStatus(ctx context.Context, resiID string) (domain.TrackingStatus, error) {
	url := fmt.Sprintf("%s/tracking/%s/status", c.baseURL, resiID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("tracking client: buat request status: %w", err)
	}
	if reqID := contextutil.GetRequestID(ctx); reqID != "" {
		req.Header.Set(contextutil.RequestIDHeader, reqID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("tracking client: kirim request status: %w", err)
	}
	defer resp.Body.Close()

	// 404 means no snapshot exists yet — resi was just created
	if resp.StatusCode == http.StatusNotFound {
		return domain.StatusCreated, nil
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("tracking client: status tidak OK: %d", resp.StatusCode)
	}

	// Decode {"success":true,"data":{"resi_id":"...","status":"..."}}
	var wrapper struct {
		Success bool `json:"success"`
		Data    struct {
			ResiID string `json:"resi_id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return "", fmt.Errorf("tracking client: decode response: %w", err)
	}

	return domain.TrackingStatus(wrapper.Data.Status), nil
}

// GetTrackingHistory mengambil semua tracking event dari Tracking Service.
// Jika paket belum punya event (404), dikembalikan slice kosong bukan error.
func (c *HTTPTrackingClient) GetTrackingHistory(ctx context.Context, resiID string) ([]domain.TrackingEvent, error) {
	url := fmt.Sprintf("%s/tracking/%s", c.baseURL, resiID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("tracking client: buat request history: %w", err)
	}
	if reqID := contextutil.GetRequestID(ctx); reqID != "" {
		req.Header.Set(contextutil.RequestIDHeader, reqID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tracking client: kirim request history: %w", err)
	}
	defer resp.Body.Close()

	// 404 means no events yet — fresh resi
	if resp.StatusCode == http.StatusNotFound {
		return []domain.TrackingEvent{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tracking client: history status tidak OK: %d", resp.StatusCode)
	}

	// Decode {"success":true,"data":[...events...]}
	var wrapper struct {
		Success bool                   `json:"success"`
		Data    []domain.TrackingEvent `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return nil, fmt.Errorf("tracking client: decode history response: %w", err)
	}
	if wrapper.Data == nil {
		return []domain.TrackingEvent{}, nil
	}
	return wrapper.Data, nil
}

// ================================================================
// HTTPPricingClient — implementasi PricingClient via HTTP
// ================================================================

// HTTPPricingClient mengirim request perhitungan harga via HTTP POST ke Pricing Service.
type HTTPPricingClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewHTTPPricingClient membuat PricingClient baru.
func NewHTTPPricingClient(baseURL string) *HTTPPricingClient {
	return &HTTPPricingClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// CalculatePrice mengirim request POST ke Pricing Service dan mengembalikan hasilnya.
func (c *HTTPPricingClient) CalculatePrice(ctx context.Context, req domain.PricingRequest) (*domain.PricingResult, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("pricing client: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/pricing", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("pricing client: buat request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if reqID := contextutil.GetRequestID(ctx); reqID != "" {
		httpReq.Header.Set(contextutil.RequestIDHeader, reqID)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("pricing client: kirim request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pricing client: status tidak OK: %d", resp.StatusCode)
	}

	var result domain.PricingResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("pricing client: decode response: %w", err)
	}

	return &result, nil
}

// ================================================================
// HTTPOrderClient — implementasi OrderClient via HTTP
// ================================================================

// HTTPOrderClient memvalidasi resi via HTTP GET ke Order Service.
type HTTPOrderClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewHTTPOrderClient membuat OrderClient baru.
func NewHTTPOrderClient(baseURL string) *HTTPOrderClient {
	return &HTTPOrderClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// ValidateResi memvalidasi resi via GET /orders/{resiID}/validate.
func (c *HTTPOrderClient) ValidateResi(ctx context.Context, resiID string) error {
	url := fmt.Sprintf("%s/orders/%s/validate", c.baseURL, resiID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("order client: buat request: %w", err)
	}
	if reqID := contextutil.GetRequestID(ctx); reqID != "" {
		req.Header.Set(contextutil.RequestIDHeader, reqID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("order client: kirim request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return domain.ErrResiInvalid
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("order client: status tidak OK: %d", resp.StatusCode)
	}

	return nil
}

// ================================================================
// HTTPAuthClient — implementasi AuthClient via HTTP
// ================================================================

// HTTPAuthClient memvalidasi JWT token via HTTP ke Auth Service.
type HTTPAuthClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewHTTPAuthClient membuat AuthClient baru.
func NewHTTPAuthClient(baseURL string) *HTTPAuthClient {
	return &HTTPAuthClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// validateRequest adalah request body untuk validasi token.
type validateRequest struct {
	Token string `json:"token"`
}

// validateResponse adalah response dari Auth Service.
type validateResponse struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

// ValidateToken memvalidasi token JWT dan mengembalikan klaim.
func (c *HTTPAuthClient) ValidateToken(ctx context.Context, token string) (*domain.AuthClaims, error) {
	body, _ := json.Marshal(validateRequest{Token: token})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/auth/validate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("auth client: buat request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if reqID := contextutil.GetRequestID(ctx); reqID != "" {
		req.Header.Set(contextutil.RequestIDHeader, reqID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("auth client: kirim request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, domain.ErrUnauthorized
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth client: status tidak OK: %d", resp.StatusCode)
	}

	var result validateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("auth client: decode response: %w", err)
	}

	return &domain.AuthClaims{
		UserID: result.UserID,
		Role:   domain.UserRole(result.Role),
	}, nil
}

// validateUserRoleRequest adalah request body untuk validasi user dan role.
type validateUserRoleRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

// ValidateUserRole memvalidasi bahwa user exists dan memiliki role yang sesuai.
// Endpoint auth service: POST /auth/validate-user-role
// Request: {"user_id": "123", "role": "courier"}
// Response 200: User valid dengan role sesuai
// Response 400/404: User tidak ditemukan atau role tidak sesuai
func (c *HTTPAuthClient) ValidateUserRole(ctx context.Context, userID string, expectedRole domain.UserRole) error {
	reqBody := validateUserRoleRequest{
		UserID: userID,
		Role:   string(expectedRole),
	}
	
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("auth client: marshal request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/auth/validate-user-role", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("auth client: buat request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if reqID := contextutil.GetRequestID(ctx); reqID != "" {
		req.Header.Set(contextutil.RequestIDHeader, reqID)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("auth client: kirim request: %w", err)
	}
	defer resp.Body.Close()
	
	// 400/404 = user tidak ditemukan atau role tidak sesuai
	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound {
		return domain.ErrInvalidCourierUser
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth client: status tidak OK: %d", resp.StatusCode)
	}
	
	return nil
}
