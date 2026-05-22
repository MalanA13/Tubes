package pricing

import (
	"errors"
	"github.com/tubes-cc/logistics/domain"
)

const (
	WeightRatePerKg      = 10000.0
	DistanceRatePerKm    = 500.0
	DeliveryPriceFloor   = 15000.0 // Minimum cost logic as requested
)

// PricingService handles business logic for cost calculations
type PricingService struct {
	repo PricingRepository
}

// NewPricingService creates a new instance of PricingService
func NewPricingService(repo PricingRepository) *PricingService {
	return &PricingService{repo: repo}
}

// CalculatePrice calculates the total shipping cost based on distance, weight, and service type
func (s *PricingService) CalculatePrice(req domain.PricingRequest) (*domain.PricingResult, error) {
	if req.Weight <= 0 {
		return nil, errors.New("weight must be greater than zero")
	}

	baseTariff, err := s.repo.GetBaseTariff(req.Origin, req.Destination, req.ServiceType)
	if err != nil {
		return nil, err
	}

	// Dynamic calculation based on weight and distance
	costByWeight := req.Weight * WeightRatePerKg
	costByDistance := req.Distance * DistanceRatePerKm
	
	totalCost := baseTariff + costByWeight + costByDistance

	// Apply Delivery Price Floor
	if totalCost < DeliveryPriceFloor {
		totalCost = DeliveryPriceFloor
	}

	return &domain.PricingResult{
		TotalCost: totalCost,
		BaseCost:  baseTariff,
	}, nil
}
