package pricing

import (
	"errors"
	"github.com/tubes-cc/logistics/domain"
	"gorm.io/gorm"
)

// Tariff represents the database model for price tariffs.
type Tariff struct {
	ID          uint               `gorm:"primaryKey"`
	Origin      string             `gorm:"index"`
	Destination string             `gorm:"index"`
	ServiceType domain.ServiceType `gorm:"index"`
	BaseCost    float64
}

// PricingRepository defines the interface for database operations related to pricing
type PricingRepository interface {
	GetBaseTariff(origin string, destination string, serviceType domain.ServiceType) (float64, error)
}

// pricingRepository is the implementation of PricingRepository using GORM
type pricingRepository struct {
	db *gorm.DB
}

// NewPricingRepository creates a new instance of PricingRepository
func NewPricingRepository(db *gorm.DB) PricingRepository {
	return &pricingRepository{db: db}
}

// GetBaseTariff fetches the tariff from DB.
func (r *pricingRepository) GetBaseTariff(origin string, destination string, serviceType domain.ServiceType) (float64, error) {
	var tariff Tariff
	result := r.db.Where("origin = ? AND destination = ? AND service_type = ?", origin, destination, serviceType).First(&tariff)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return 0, errors.New("tariff not found")
		}
		return 0, result.Error
	}
	return tariff.BaseCost, nil
}
