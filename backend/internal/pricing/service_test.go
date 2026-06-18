package pricing

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tubes-cc/logistics/domain"
	"go.uber.org/mock/gomock"
)

func TestPricingService_CalculatePrice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockPricingRepository(ctrl)
	service := NewPricingService(mockRepo)

	t.Run("Normal Calculation - Express", func(t *testing.T) {
		mockRepo.EXPECT().GetBaseTariff("Jakarta", "Bandung", domain.ServiceExpress).Return(10000.0, nil)

		req := domain.PricingRequest{
			Origin:      "Jakarta",
			Destination: "Bandung",
			Weight:      2.5, // 2.5 * 10000 = 25000
			Distance:    150, // 150 * 500 = 75000
			ServiceType: domain.ServiceExpress,
		}

		res, err := service.CalculatePrice(req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		// totalCost = baseTariff (10000) + costByWeight (25000) + costByDistance (75000) = 110000
		assert.Equal(t, 110000.0, res.TotalCost)
		assert.Equal(t, 10000.0, res.BaseCost)
	})

	t.Run("Delivery Price Floor applied", func(t *testing.T) {
		mockRepo.EXPECT().GetBaseTariff("Jakarta", "Jakarta", domain.ServiceSameday).Return(2000.0, nil)

		req := domain.PricingRequest{
			Origin:      "Jakarta",
			Destination: "Jakarta",
			Weight:      0.1, // 0.1 * 10000 = 1000
			Distance:    5,   // 5 * 500 = 2500
			ServiceType: domain.ServiceSameday,
		}

		res, err := service.CalculatePrice(req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		// totalCost = 2000 + 1000 + 2500 = 5500 < 15000, so it applies DeliveryPriceFloor
		assert.Equal(t, 15000.0, res.TotalCost)
		assert.Equal(t, 2000.0, res.BaseCost)
	})

	t.Run("Negative or Zero Weight Error", func(t *testing.T) {
		req := domain.PricingRequest{
			Origin:      "Jakarta",
			Destination: "Bandung",
			Weight:      0,
			Distance:    150,
			ServiceType: domain.ServiceRegular,
		}

		// GetBaseTariff should not be called
		res, err := service.CalculatePrice(req)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "weight must be greater than zero", err.Error())
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo.EXPECT().GetBaseTariff("Unknown", "Unknown", domain.ServiceNextday).Return(0.0, errors.New("tariff not found"))

		req := domain.PricingRequest{
			Origin:      "Unknown",
			Destination: "Unknown",
			Weight:      1.0,
			Distance:    10.0,
			ServiceType: domain.ServiceNextday,
		}

		res, err := service.CalculatePrice(req)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "tariff not found", err.Error())
	})
}
