// Package domain — definisi error khusus yang digunakan di seluruh aplikasi.
package domain

import "errors"

var (
	// ErrShipmentNotFound dikembalikan ketika resi tidak ditemukan.
	ErrShipmentNotFound = errors.New("shipment not found")

	// ErrResiInvalid dikembalikan ketika OrderClient menolak resi (tidak ada / sudah batal).
	ErrResiInvalid = errors.New("resi is invalid or does not exist")

	// ErrInvalidStatus dikembalikan ketika transisi status tidak valid.
	ErrInvalidStatus = errors.New("invalid status transition")

	// ErrAlreadyScannedIn dikembalikan ketika paket sudah di-scan-in di hub yang sama.
	ErrAlreadyScannedIn = errors.New("shipment already scanned in at this hub")

	// ErrCourierAlreadyAssigned dikembalikan ketika kurir sudah ditugaskan.
	ErrCourierAlreadyAssigned = errors.New("courier already assigned to this shipment")

	// ErrInvalidResiID dikembalikan ketika resiID kosong.
	ErrInvalidResiID = errors.New("resi ID is required")

	// ErrInvalidHubID dikembalikan ketika hubID kosong.
	ErrInvalidHubID = errors.New("hub ID is required")

	// ErrInvalidCourierID dikembalikan ketika courierID kosong.
	ErrInvalidCourierID = errors.New("courier ID is required")

	// ErrInvalidProofURL dikembalikan ketika proof URL kosong saat status DELIVERED.
	ErrInvalidProofURL = errors.New("proof URL is required for DELIVERED status")

	// ErrUnauthorized dikembalikan ketika token tidak valid atau kadaluarsa.
	ErrUnauthorized = errors.New("unauthorized: invalid or missing token")

	// ErrForbidden dikembalikan ketika role tidak memiliki akses.
	ErrForbidden = errors.New("forbidden: insufficient role")
)
