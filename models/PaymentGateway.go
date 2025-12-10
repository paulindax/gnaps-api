package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"time"
)

// PaymentGateway model generated from database table 'payment_gateways'
type PaymentGateway struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	Name              *string         `json:"name" gorm:"column:name"`
	GatewayType       *string         `json:"gateway_type" gorm:"column:gateway_type"`
	GatewayParameters *datatypes.JSON `json:"gateway_parameters" gorm:"column:gateway_parameters"`
	TransactionFee    *float64        `json:"transaction_fee" gorm:"column:transaction_fee"`
	IsDeleted         *bool           `json:"is_deleted" gorm:"column:is_deleted"`
}

func (PaymentGateway) TableName() string {
	return "payment_gateways"
}
