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
