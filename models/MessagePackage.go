package models

import (
	"time"
)

// MessagePackage model generated from database table 'message_packages'
type MessagePackage struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	OwnerId   *int      `json:"owner_id" gorm:"column:owner_id"`

	OwnerType  *string `json:"owner_type" gorm:"column:owner_type"`
	Sendername *string `json:"sendername" gorm:"column:sendername"`
	Units      *int    `json:"units" gorm:"column:units"`
}

func (MessagePackage) TableName() string {
	return "message_packages"
}
