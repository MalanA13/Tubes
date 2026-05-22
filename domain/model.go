// Package domain berisi model-model data inti (entities) yang digunakan
// di seluruh microservice logistik.
// Tidak bergantung pada package lain — innermost layer dalam clean architecture.
package domain

import "time"

// TrackingStatus merepresentasikan status pengiriman paket.
type TrackingStatus string

const (
	StatusCreated     TrackingStatus = "CREATED"      // Resi baru dibuat
	StatusInHub       TrackingStatus = "IN_HUB"       // Paket sudah masuk hub (scan-in)
	StatusInTransit   TrackingStatus = "IN_TRANSIT"   // Paket keluar hub (scan-out)
	StatusOutDelivery TrackingStatus = "OUT_DELIVERY" // Paket sedang diantar kurir
	StatusDelivered   TrackingStatus = "DELIVERED"    // Paket berhasil dikirim
	StatusFailed      TrackingStatus = "FAILED"       // Pengiriman gagal
	StatusReturned    TrackingStatus = "RETURNED"     // Paket dikembalikan
)

// ServiceType merepresentasikan jenis layanan pengiriman.
type ServiceType string

const (
	ServiceRegular ServiceType = "REGULAR"
	ServiceExpress ServiceType = "EXPRESS"
	ServiceSameday ServiceType = "SAMEDAY"
	ServiceNextday ServiceType = "NEXTDAY"
)

// PricingRequest merepresentasikan input untuk perhitungan harga.
type PricingRequest struct {
	Origin      string      `json:"origin"`
	Destination string      `json:"destination"`
	Weight      float64     `json:"weight"` // in Kg
	Distance    float64     `json:"distance"` // in Km
	ServiceType ServiceType `json:"service_type"`
}

// PricingResult merepresentasikan hasil perhitungan harga.
type PricingResult struct {
	TotalCost float64 `json:"total_cost"`
	BaseCost  float64 `json:"base_cost"`
}

// Shipment merepresentasikan data pengiriman / resi.
type Shipment struct {
	ResiID    string         `json:"resi_id"`
	Status    TrackingStatus `json:"status"`
	HubID     string         `json:"hub_id"`
	CourierID string         `json:"courier_id"`
	ProofURL  string         `json:"proof_url"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// TrackingEvent menyimpan history pergerakan paket.
type TrackingEvent struct {
	ID        string         `json:"id"`        // UUID — harus unik
	ResiID    string         `json:"resi_id"`
	Status    TrackingStatus `json:"status"`
	Location  string         `json:"location"`  // Nama hub / ID kurir
	Note      string         `json:"note"`
	CreatedAt time.Time      `json:"created_at"`
}

// UserRole merepresentasikan peran pengguna untuk otorisasi.
type UserRole string

const (
	RoleAdmin   UserRole = "admin"
	RoleCourier UserRole = "courier"
)

// AuthClaims merepresentasikan klaim JWT yang sudah divalidasi.
type AuthClaims struct {
	UserID string   `json:"user_id"`
	Role   UserRole `json:"role"`
}
