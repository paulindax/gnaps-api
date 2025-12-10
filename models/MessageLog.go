package models

import (
	"gorm.io/datatypes"
	"time"
)

// MessageLog model generated from database table 'message_logs'
type MessageLog struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	MsgType   *string   `json:"msg_type" gorm:"column:msg_type"`

	Message         *string         `json:"message" gorm:"column:message"`
	Recipient       *string         `json:"recipient" gorm:"column:recipient"`
	OwnerType       *string         `json:"owner_type" gorm:"column:owner_type"`
	OwnerId         *int64          `json:"owner_id" gorm:"column:owner_id"`
	GatewayResponse *string         `json:"gateway_response" gorm:"column:gateway_response"`
	Notify          *bool           `json:"notify" gorm:"column:notify"`
	Units           *float64        `json:"units" gorm:"column:units"`
	UserIds         *datatypes.JSON `json:"user_ids" gorm:"column:user_ids"`
}

func (MessageLog) TableName() string {
	return "message_logs"
}
