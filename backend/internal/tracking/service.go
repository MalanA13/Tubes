package tracking

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/tubes-cc/logistics/domain"
)

// TrackingService handles business logic for package tracking
type TrackingService struct {
	repo TrackingRepository
}

// NewTrackingService creates a new instance of TrackingService
func NewTrackingService(repo TrackingRepository) *TrackingService {
	return &TrackingService{repo: repo}
}

// AddTrackingEvent creates a new event and updates the overall shipment status
func (s *TrackingService) AddTrackingEvent(resiID string, status domain.TrackingStatus, location string, note string) error {
	if resiID == "" {
		return errors.New("resi_id cannot be empty")
	}

	event := domain.TrackingEvent{
		ID:        uuid.NewString(),
		ResiID:    resiID,
		Status:    status,
		Location:  location,
		Note:      note,
		CreatedAt: time.Now(),
	}

	err := s.repo.SaveEvent(event)
	if err != nil {
		return err
	}

	return s.repo.UpdateShipmentStatus(resiID, status)
}

// GetHistory retrieves the tracking history of a given resiID
func (s *TrackingService) GetHistory(resiID string) ([]domain.TrackingEvent, error) {
	if resiID == "" {
		return nil, errors.New("resi_id cannot be empty")
	}
	return s.repo.GetTrackingHistory(resiID)
}

// GetCurrentStatus returns the current shipment status from the tracking snapshot.
// Returns domain.ErrShipmentNotFound if no snapshot exists for the resi.
func (s *TrackingService) GetCurrentStatus(resiID string) (domain.TrackingStatus, error) {
	if resiID == "" {
		return "", errors.New("resi_id cannot be empty")
	}
	return s.repo.GetCurrentStatus(resiID)
}

// ListAllShipments returns all shipments in the system for admin use.
// No user filtering is applied.
func (s *TrackingService) ListAllShipments() ([]ShipmentModel, error) {
	return s.repo.ListAllShipments()
}
