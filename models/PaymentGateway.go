package models

import (
	"time"
	"gorm.io/datatypes"
)

// PaymentGateway model generated from database table 'payment_gateways'
type PaymentGateway struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	Name *string `json:"name" gorm:"column:name"`

	GatewayType *string `json:"gateway_type" gorm:"column:gateway_type"`
	ApiLiveLink *string `json:"api_live_link" gorm:"column:api_live_link"`
	ApiTestLink *string `json:"api_test_link" gorm:"column:api_test_link"`
	GatewayParameters *datatypes.JSON `json:"gateway_parameters" gorm:"column:gateway_parameters"`
	TransactionFee *float64 `json:"transaction_fee" gorm:"column:transaction_fee"`
	PaymentInstructions *string `json:"payment_instructions" gorm:"column:payment_instructions"`
	IsDeleted *bool `json:"is_deleted" gorm:"column:is_deleted"`
	SchoolCode *string `json:"school_code" gorm:"column:school_code"`
	SchoolId *int64 `json:"school_id" gorm:"column:school_id"`
}

func (PaymentGateway) TableName() string {
	return "payment_gateways"
}
