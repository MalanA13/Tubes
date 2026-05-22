package order

import (
	"time"

	"github.com/tubes-cc/logistics/domain"
	"gorm.io/gorm"
)

// OrderModel represents the order in the database.
type OrderModel struct {
	OrderID       string `gorm:"primaryKey"`
	ResiID        string `gorm:"uniqueIndex"`
	SenderName    string
	RecipientName string
	Origin        string
	Destination   string
	Weight        float64
	ItemType      string
	ServiceType   domain.ServiceType
	TotalCost     float64
	Status        domain.TrackingStatus
	CreatedAt     time.Time
}

// OrderRepository defines the interface for order DB operations.
type OrderRepository interface {
	SaveOrder(order *OrderModel) error
}

type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a new instance.
func NewOrderRepository(db *gorm.DB) (OrderRepository, error) {
	err := db.AutoMigrate(&OrderModel{})
	if err != nil {
		return nil, err
	}
	return &orderRepository{db: db}, nil
}

// SaveOrder saves an order into the database.
func (r *orderRepository) SaveOrder(order *OrderModel) error {
	return r.db.Create(order).Error
}
