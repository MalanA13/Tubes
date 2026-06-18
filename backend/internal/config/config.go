package config

import "os"

// GetEnv helper method to read environment variables with defaults.
func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// AuthConfig holds configurations for the Auth service.
type AuthConfig struct {
	Port   string
	DBPath string
}

// LoadAuthConfig loads configuration for the Auth service.
func LoadAuthConfig() *AuthConfig {
	return &AuthConfig{
		Port:   GetEnv("PORT", "8080"),
		DBPath: GetEnv("AUTH_DB", "auth.db"),
	}
}

// OrderConfig holds configurations for the Order service.
type OrderConfig struct {
	Port           string
	DBPath         string
	TrackingSvcURL string
	PricingSvcURL  string
	AuthSvcURL     string
}

// LoadOrderConfig loads configuration for the Order service.
func LoadOrderConfig() *OrderConfig {
	return &OrderConfig{
		Port:           GetEnv("PORT", "8081"),
		DBPath:         GetEnv("ORDER_DB", "order.db"),
		TrackingSvcURL: GetEnv("TRACKING_SVC_URL", "http://tracking-service:8083"),
		PricingSvcURL:  GetEnv("PRICING_SVC_URL", "http://pricing-service:8082"),
		AuthSvcURL:     GetEnv("AUTH_SVC_URL", "http://auth-service:8080"),
	}
}

// PricingConfig holds configurations for the Pricing service.
type PricingConfig struct {
	Port   string
	DBPath string
}

// LoadPricingConfig loads configuration for the Pricing service.
func LoadPricingConfig() *PricingConfig {
	return &PricingConfig{
		Port:   GetEnv("PORT", "8082"),
		DBPath: GetEnv("PRICING_DB", "pricing.db"),
	}
}

// TrackingConfig holds configurations for the Tracking service.
type TrackingConfig struct {
	Port       string
	DBPath     string
	AuthSvcURL string
}

// LoadTrackingConfig loads configuration for the Tracking service.
func LoadTrackingConfig() *TrackingConfig {
	return &TrackingConfig{
		Port:       GetEnv("PORT", "8083"),
		DBPath:     GetEnv("TRACKING_DB", "tracking.db"),
		AuthSvcURL: GetEnv("AUTH_SVC_URL", "http://auth-service:8080"),
	}
}

// HubConfig holds configurations for the Hub service.
type HubConfig struct {
	Port           string
	DBPath         string
	TrackingSvcURL string
	OrderSvcURL    string
	AuthSvcURL     string
}

// LoadHubConfig loads configuration for the Hub service.
func LoadHubConfig() *HubConfig {
	return &HubConfig{
		Port:           GetEnv("PORT", "8084"),
		DBPath:         GetEnv("HUB_DB", "hub.db"),
		TrackingSvcURL: GetEnv("TRACKING_SVC_URL", "http://tracking-service:8083"),
		OrderSvcURL:    GetEnv("ORDER_SVC_URL", "http://order-service:8081"),
		AuthSvcURL:     GetEnv("AUTH_SVC_URL", "http://auth-service:8080"),
	}
}

// CourierConfig holds configurations for the Courier service.
type CourierConfig struct {
	Port           string
	DBPath         string
	TrackingSvcURL string
	OrderSvcURL    string
	AuthSvcURL     string
}

// LoadCourierConfig loads configuration for the Courier service.
func LoadCourierConfig() *CourierConfig {
	return &CourierConfig{
		Port:           GetEnv("PORT", "8085"),
		DBPath:         GetEnv("COURIER_DB", "courier.db"),
		TrackingSvcURL: GetEnv("TRACKING_SVC_URL", "http://tracking-service:8083"),
		OrderSvcURL:    GetEnv("ORDER_SVC_URL", "http://order-service:8081"),
		AuthSvcURL:     GetEnv("AUTH_SVC_URL", "http://auth-service:8080"),
	}
}

// NotificationConfig holds configurations for the Notification service.
type NotificationConfig struct {
	Port   string
	DBPath string
}

// LoadNotificationConfig loads configuration for the Notification service.
func LoadNotificationConfig() *NotificationConfig {
	return &NotificationConfig{
		Port:   GetEnv("PORT", "8080"),
		DBPath: GetEnv("NOTIFICATION_DB", "notification.db"),
	}
}
