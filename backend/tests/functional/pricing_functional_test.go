package functional

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tubes-cc/logistics/domain"
	"github.com/tubes-cc/logistics/internal/pricing"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPricingFunctional(t *testing.T) {
	// Setup in-memory SQLite using GORM
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	// Migrate the tariff model
	err = db.AutoMigrate(&pricing.Tariff{})
	require.NoError(t, err)

	// Seed Dummy Data
	db.Create(&pricing.Tariff{Origin: "Jakarta", Destination: "Bandung", ServiceType: domain.ServiceExpress, BaseCost: 10000.0})
	db.Create(&pricing.Tariff{Origin: "Jakarta", Destination: "Jakarta", ServiceType: domain.ServiceSameday, BaseCost: 2000.0})

	// Setup Repo & Service
	repo := pricing.NewPricingRepository(db)
	service := pricing.NewPricingService(repo)

	t.Run("Normal Calculation Functional - Express", func(t *testing.T) {
		req := domain.PricingRequest{
			Origin:      "Jakarta",
			Destination: "Bandung",
			Weight:      2.5, // 25000
			Distance:    150, // 75000
			ServiceType: domain.ServiceExpress,
		}

		res, err := service.CalculatePrice(req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, 110000.0, res.TotalCost)
		assert.Equal(t, 10000.0, res.BaseCost)
	})

	t.Run("Delivery Price Floor Functional", func(t *testing.T) {
		req := domain.PricingRequest{
			Origin:      "Jakarta",
			Destination: "Jakarta",
			Weight:      0.1, // 1000
			Distance:    5,   // 2500
			ServiceType: domain.ServiceSameday,
		}

		res, err := service.CalculatePrice(req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, pricing.DeliveryPriceFloor, res.TotalCost)
		assert.Equal(t, 2000.0, res.BaseCost)
	})
}
