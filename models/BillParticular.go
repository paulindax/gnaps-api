package models

import (
	"gorm.io/gorm"
	"time"
)

// BillParticular model generated from database table 'bill_particulars'
type BillParticular struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	Name *string `json:"name" gorm:"column:name"`
	Priority *int `json:"priority" gorm:"column:priority"`
	FinanceAccountId *int64 `json:"finance_account_id" gorm:"column:finance_account_id"`
	IsDeleted *bool `json:"is_deleted" gorm:"column:is_deleted"`
	IsArrears *bool `json:"is_arrears" gorm:"column:is_arrears"`
}

func (BillParticular) TableName() string {
	return "bill_particulars"
}
